FROM fedora:33

COPY microshift /usr/local/bin/microshift
COPY unit /usr/lib/systemd/system/microshift.service
COPY kubelet-cgroups.conf /etc/systemd/system.conf.d/kubelet-cgroups.conf

RUN dnf module -y enable cri-o:1.20 && \
    dnf install -y cri-o \
        cri-tools \
        conntrack \
        iptables-services \ 
        iproute \
        procps-ng && \
    dnf clean all

RUN curl -LO "https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl" && \
    chmod +x ./kubectl && \
    mv ./kubectl /usr/local/bin/kubectl && \
    sed -i 's/10.85.0.0\/16/10.42.0.0\/24/' /etc/cni/net.d/100-crio-bridge.conf && \
    sed -i 's/0.3.1/0.4.0/' /etc/cni/net.d/100-crio-bridge.conf && \
    systemctl enable microshift.service && \
    systemctl enable crio 

CMD [ "/sbin/init" ]