FROM centos:8

ENV HOME=/app/

ADD natok-cli ${HOME}
ADD s-cert.key ${HOME}
ADD s-cert.pem ${HOME}
ADD application.json ${HOME}

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

WORKDIR ${HOME}
ENTRYPOINT ["sh","-c","./natok-cli"]

# docker build -f Dockerfile -t natok-cli:0.1 .
# docker run --name=natok-cli --restart=always --net=host -d natok-cli:0.1
