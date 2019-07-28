FROM alexwoo/centos-dev
MAINTAINER AlexWoo <wj19840501@gmail.com>

COPY ./ /root/work/gortc/
RUN cd /root/work/gortc && ./install
RUN rm -rf /root/work/gortc

EXPOSE 8080 2539 22
VOLUME /usr/local/gortc/logs/

CMD /usr/local/gortc/bin/gortc
