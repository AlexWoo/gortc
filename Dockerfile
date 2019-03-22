FROM alexwoo/golang as install
MAINTAINER AlexWoo <wj19840501@gmail.com>
RUN git clone https://github.com/AlexWoo/gortc.git && cd gortc && ./install

FROM centos
COPY --from=install /usr/local/gortc /usr/local/gortc
EXPOSE 8080 2539
VOLUME /usr/local/gortc/logs/
WORKDIR /usr/local/gortc
CMD /usr/local/gortc/bin/gortc
