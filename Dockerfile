FROM centos as install
MAINTAINER AlexWoo <wj19840501@gmail.com>
WORKDIR /root

# base package
RUN yum install -y make gcc git openssl epel-release
RUN yum install -y golang
RUN git clone https://github.com/AlexWoo/gortc.git && cd gortc && ./install

FROM centos
COPY --from=install /usr/local/gortc /usr/local/gortc
EXPOSE 8080 2539
VOLUME /usr/local/gortc/logs/
WORKDIR /usr/local/gortc
CMD /usr/local/gortc/bin/gortc
