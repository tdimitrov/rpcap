#Copy-pasted from https://docs.docker.com/engine/examples/running_ssh_service/#build-an-eg_sshd-image
FROM ubuntu:20.04

RUN apt-get update && apt-get install -y openssh-server tcpdump
RUN mkdir /var/run/sshd

# Authentication
ADD ssh-key.pub /root/.ssh/authorized_keys

RUN sed -i 's/PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config

# SSH login fix. Otherwise user is kicked off after login
RUN sed 's@session\s*required\s*pam_loginuid.so@session optional pam_loginuid.so@g' -i /etc/pam.d/sshd

ENV NOTVISIBLE "in users profile"
RUN echo "export VISIBLE=now" >> /etc/profile

EXPOSE 22

CMD ["/usr/sbin/sshd", "-D"]