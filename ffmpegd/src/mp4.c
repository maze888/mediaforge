#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include <libavformat/avformat.h>
#include <libavcodec/avcodec.h>
#include <libavutil/avutil.h>
#include <libavutil/opt.h>
#include <libavutil/channel_layout.h>
#include <libswresample/swresample.h> 
#include <libavutil/audio_fifo.h> 

#define CHECK_RET(ret, msg) if ((ret) < 0) { fprintf(stderr, msg " 실패: %s\n", av_err2str(ret)); goto end; }

static int open_input_file(const char *filename, AVFormatContext **input_format_context, int *audio_stream_index, AVCodecContext **input_codec_context);
static int open_output_file(const char *filename, AVFormatContext **output_format_context, AVCodecContext **output_codec_context);
static int transcode_audio(AVFormatContext *input_format_context, AVCodecContext *input_codec_context, AVFormatContext *output_format_context, AVCodecContext *output_codec_context, int audio_stream_index);

int extract_mp3(const char *input_filepath, const char *output_filepath) {
    AVFormatContext *input_format_context = NULL;
    AVFormatContext *output_format_context = NULL;
    AVCodecContext *input_codec_context = NULL;
    AVCodecContext *output_codec_context = NULL;
    int audio_stream_index = -1;
    int ret = 0;

    ret = open_input_file(input_filepath, &input_format_context, &audio_stream_index, &input_codec_context);
    CHECK_RET(ret, "입력 파일 설정");
    ret = open_output_file(output_filepath, &output_format_context, &output_codec_context);
    CHECK_RET(ret, "출력 파일 설정");
    ret = avcodec_open2(output_codec_context, avcodec_find_encoder(AV_CODEC_ID_MP3), NULL);
    CHECK_RET(ret, "출력 코덱 열기");

    ret = transcode_audio(input_format_context, input_codec_context,
                          output_format_context, output_codec_context,
                          audio_stream_index);
    CHECK_RET(ret, "오디오 트랜스코딩");

    av_write_trailer(output_format_context);
    printf("성공적으로 MP3 파일을 추출하고 저장했습니다: %s\n", output_filepath);

end:
    if (input_codec_context) avcodec_free_context(&input_codec_context);
    if (output_codec_context) avcodec_free_context(&output_codec_context);
    if (input_format_context) avformat_close_input(&input_format_context);
    if (output_format_context && !(output_format_context->oformat->flags & AVFMT_NOFILE))
        avio_closep(&output_format_context->pb);
    if (output_format_context) avformat_free_context(output_format_context);

    if (ret < 0) {
        return 1;
    }
    return 0;
}

static int open_input_file(const char *filename, AVFormatContext **input_format_context, 
                           int *audio_stream_index, AVCodecContext **input_codec_context) {
    int ret;
    ret = avformat_open_input(input_format_context, filename, NULL, NULL);
    CHECK_RET(ret, "입력 파일을 열 수 없습니다");
    ret = avformat_find_stream_info(*input_format_context, NULL);
    CHECK_RET(ret, "스트림 정보를 찾을 수 없습니다");
    *audio_stream_index = av_find_best_stream(*input_format_context, AVMEDIA_TYPE_AUDIO, -1, -1, NULL, 0);
    if (*audio_stream_index < 0) {
        fprintf(stderr, "입력 파일에서 오디오 스트림을 찾을 수 없습니다.\n");
        ret = AVERROR_STREAM_NOT_FOUND; goto end;
    }
    AVStream *in_stream = (*input_format_context)->streams[*audio_stream_index];
    const AVCodec *input_codec = avcodec_find_decoder(in_stream->codecpar->codec_id);
    if (!input_codec) { fprintf(stderr, "입력 오디오 디코더를 찾을 수 없습니다.\n"); ret = AVERROR_DECODER_NOT_FOUND; goto end; }
    *input_codec_context = avcodec_alloc_context3(input_codec);
    if (!*input_codec_context) { fprintf(stderr, "입력 코덱 컨텍스트 할당에 실패했습니다.\n"); ret = AVERROR(ENOMEM); goto end; }
    avcodec_parameters_to_context(*input_codec_context, in_stream->codecpar);
    ret = avcodec_open2(*input_codec_context, input_codec, NULL);
    CHECK_RET(ret, "입력 코덱 열기");
    return 0;
end:
    if (*input_format_context) avformat_close_input(input_format_context);
    if (*input_codec_context) avcodec_free_context(input_codec_context);
    return ret;
}

static int open_output_file(const char *filename, AVFormatContext **output_format_context, AVCodecContext **output_codec_context) {
    int ret;
    const AVCodec *output_codec = avcodec_find_encoder(AV_CODEC_ID_MP3);
    if (!output_codec) { fprintf(stderr, "MP3 인코더(libmp3lame)를 찾을 수 없습니다.\n"); return AVERROR_ENCODER_NOT_FOUND; }
    ret = avformat_alloc_output_context2(output_format_context, NULL, "mp3", filename);
    CHECK_RET(ret, "출력 컨텍스트 할당");
    AVStream *out_stream = avformat_new_stream(*output_format_context, output_codec);
    if (!out_stream) { fprintf(stderr, "출력 스트림 생성에 실패했습니다.\n"); ret = AVERROR(ENOMEM); goto end; }
    *output_codec_context = avcodec_alloc_context3(output_codec);
    if (!*output_codec_context) { fprintf(stderr, "출력 코덱 컨텍스트 할당에 실패했습니다.\n"); ret = AVERROR(ENOMEM); goto end; }
    AVChannelLayout out_ch_layout = AV_CHANNEL_LAYOUT_STEREO;
    (*output_codec_context)->sample_rate    = 44100;
    (*output_codec_context)->bit_rate       = 320000; //192000;
    (*output_codec_context)->sample_fmt     = output_codec->sample_fmts[0]; 
    (*output_codec_context)->time_base      = (AVRational){1, (*output_codec_context)->sample_rate};
    av_channel_layout_copy(&(*output_codec_context)->ch_layout, &out_ch_layout);
    ret = avcodec_parameters_from_context(out_stream->codecpar, *output_codec_context);
    CHECK_RET(ret, "코덱 매개변수 복사");
    if (!((*output_format_context)->oformat->flags & AVFMT_NOFILE)) {
        ret = avio_open(&(*output_format_context)->pb, filename, AVIO_FLAG_WRITE);
        CHECK_RET(ret, "출력 파일 열기");
    }
    ret = avformat_write_header(*output_format_context, NULL);
    CHECK_RET(ret, "헤더 작성");
    av_dump_format(*output_format_context, 0, filename, 1);
    return 0;
end:
    if (*output_codec_context) avcodec_free_context(output_codec_context);
    if (*output_format_context) avformat_free_context(*output_format_context);
    return ret;
}

// 오디오 트랜스코딩 과정 (FIFO 버퍼링 포함)
static int transcode_audio(AVFormatContext *input_format_context, AVCodecContext *input_codec_context,
                           AVFormatContext *output_format_context, AVCodecContext *output_codec_context,
                           int audio_stream_index) {

    int ret = 0;
    AVPacket input_packet = {0};
    AVFrame *decoded_frame = av_frame_alloc();
    AVFrame *resampled_frame = av_frame_alloc(); 
    AVFrame *output_frame = av_frame_alloc(); 
    AVPacket output_packet = {0};
    struct SwrContext *swr_ctx = NULL;
    AVAudioFifo *audio_fifo = NULL; 
    AVChannelLayout in_ch_layout = {0};

    if (!decoded_frame || !resampled_frame || !output_frame) { ret = AVERROR(ENOMEM); goto end; }

    // 1. SwrContext 설정 및 FIFO 초기화
    AVCodecParameters *codecpar = input_format_context->streams[audio_stream_index]->codecpar;
    ret = av_channel_layout_copy(&in_ch_layout, &codecpar->ch_layout);
    CHECK_RET(ret, "입력 채널 레이아웃 복사");
    
    ret = swr_alloc_set_opts2(&swr_ctx,
                              &output_codec_context->ch_layout, output_codec_context->sample_fmt, output_codec_context->sample_rate,
                              &in_ch_layout, input_codec_context->sample_fmt, input_codec_context->sample_rate,
                              0, NULL);
    CHECK_RET(ret, "SwrContext 설정");
    ret = swr_init(swr_ctx);
    CHECK_RET(ret, "SwrContext 초기화");

    audio_fifo = av_audio_fifo_alloc(output_codec_context->sample_fmt, 
                                     output_codec_context->ch_layout.nb_channels, 
                                     output_codec_context->frame_size);
    if (!audio_fifo) { ret = AVERROR(ENOMEM); goto end; }

    // 인코딩할 최종 프레임(output_frame) 속성 설정 및 버퍼 확보
    av_channel_layout_copy(&output_frame->ch_layout, &output_codec_context->ch_layout);
    output_frame->sample_rate = output_codec_context->sample_rate;
    output_frame->format = output_codec_context->sample_fmt;
    output_frame->nb_samples = output_codec_context->frame_size; 
    ret = av_frame_get_buffer(output_frame, 0);
    CHECK_RET(ret, "출력 프레임 버퍼 할당"); // <--- 이 버퍼를 루프 내에서 계속 재사용합니다.

    // 2. 메인 루프: 읽기, 디코딩, 리샘플링, FIFO 저장
    while (av_read_frame(input_format_context, &input_packet) >= 0 || 
           av_audio_fifo_size(audio_fifo) >= output_codec_context->frame_size) {
        
        if (input_packet.stream_index == audio_stream_index) {
            
            // A. 디코딩
            ret = avcodec_send_packet(input_codec_context, &input_packet);
            if (ret >= 0) { 
                while (avcodec_receive_frame(input_codec_context, decoded_frame) == 0) {
                    
                    // B. 리샘플링
                    av_channel_layout_copy(&resampled_frame->ch_layout, &output_codec_context->ch_layout);
                    resampled_frame->sample_rate = output_codec_context->sample_rate;
                    resampled_frame->format = output_codec_context->sample_fmt;
                    resampled_frame->nb_samples = 0; 
                    
                    ret = swr_convert_frame(swr_ctx, resampled_frame, decoded_frame);
                    CHECK_RET(ret, "리샘플링 오류");

                    // C. FIFO에 쓰기
                    if (resampled_frame->nb_samples > 0) {
                        ret = av_audio_fifo_write(audio_fifo, (void**)resampled_frame->data, resampled_frame->nb_samples);
                        if (ret < resampled_frame->nb_samples) {
                            fprintf(stderr, "FIFO 쓰기 실패\n");
                            ret = AVERROR_UNKNOWN;
                            goto end;
                        }
                    }

                    av_frame_unref(decoded_frame);
                    av_frame_unref(resampled_frame); 
                }
            }
        }
        av_packet_unref(&input_packet);

        // 3. FIFO에서 읽어와 인코딩 (FIFO에 1152 샘플 이상이 쌓였을 경우)
        while (av_audio_fifo_size(audio_fifo) >= output_codec_context->frame_size) {
            
            output_frame->nb_samples = output_codec_context->frame_size;
            ret = av_audio_fifo_read(audio_fifo, (void**)output_frame->data, output_frame->nb_samples);
            if (ret < output_frame->nb_samples) {
                fprintf(stderr, "FIFO 읽기 실패\n");
                ret = AVERROR_UNKNOWN;
                goto end;
            }
            
            // **[SIGSEGV 해결]** output_frame은 여기서 unref 하지 않습니다. 데이터 포인터를 재사용합니다.

            // D. 인코딩 및 패킷 쓰기
            ret = avcodec_send_frame(output_codec_context, output_frame);
            CHECK_RET(ret, "프레임 전송 오류");

            while (avcodec_receive_packet(output_codec_context, &output_packet) == 0) {
                av_packet_rescale_ts(&output_packet, 
                                     input_format_context->streams[audio_stream_index]->time_base,
                                     output_format_context->streams[0]->time_base);
                output_packet.stream_index = 0; 
                av_write_frame(output_format_context, &output_packet);
                av_packet_unref(&output_packet);
            }
            // output_frame에 대한 av_frame_unref(output_frame) 호출은 **제거**되었습니다.
        }
    }

    // 4. 플러시(Flush) 처리: 잔여 FIFO 데이터 및 인코더 플러시
    
    // 4.1. SwrContext 플러시
    do {
        ret = swr_convert_frame(swr_ctx, resampled_frame, NULL);
        if (ret < 0) break; 
        if (resampled_frame->nb_samples > 0) {
            av_audio_fifo_write(audio_fifo, (void**)resampled_frame->data, resampled_frame->nb_samples);
        }
        av_frame_unref(resampled_frame);
    } while (ret > 0);

    // 4.2. FIFO의 잔여 데이터를 인코딩
    while (av_audio_fifo_size(audio_fifo) > 0) {
        output_frame->nb_samples = FFMIN(av_audio_fifo_size(audio_fifo), output_codec_context->frame_size); 
        ret = av_audio_fifo_read(audio_fifo, (void**)output_frame->data, output_frame->nb_samples);
        if (ret < output_frame->nb_samples) {
            fprintf(stderr, "최종 FIFO 읽기 실패\n");
            ret = AVERROR_UNKNOWN;
            goto end;
        }

        ret = avcodec_send_frame(output_codec_context, output_frame);
        if (ret < 0) break;
        
        while (avcodec_receive_packet(output_codec_context, &output_packet) == 0) {
            av_packet_rescale_ts(&output_packet, input_format_context->streams[audio_stream_index]->time_base,
                                 output_format_context->streams[0]->time_base);
            output_packet.stream_index = 0;
            av_write_frame(output_format_context, &output_packet);
            av_packet_unref(&output_packet);
        }
        // **[SIGSEGV 해결]** output_frame은 여기서 unref 하지 않습니다.
    }

    // 4.3. 인코더 플러시
    ret = avcodec_send_frame(output_codec_context, NULL); 
    if (ret >= 0) { 
        while (avcodec_receive_packet(output_codec_context, &output_packet) == 0) {
            av_packet_rescale_ts(&output_packet, input_format_context->streams[audio_stream_index]->time_base,
                                 output_format_context->streams[0]->time_base);
            output_packet.stream_index = 0;
            av_write_frame(output_format_context, &output_packet);
            av_packet_unref(&output_packet);
        }
    }

    ret = 0; 

end:
    // 자원 해제
    av_packet_unref(&input_packet);
    av_packet_unref(&output_packet);
    if (decoded_frame) av_frame_free(&decoded_frame);
    if (resampled_frame) av_frame_free(&resampled_frame);
    if (output_frame) av_frame_free(&output_frame); // <--- 여기서만 최종 해제!
    if (swr_ctx) swr_free(&swr_ctx);
    if (audio_fifo) av_audio_fifo_free(audio_fifo);
    av_channel_layout_uninit(&in_ch_layout);
    
    return ret;
}
