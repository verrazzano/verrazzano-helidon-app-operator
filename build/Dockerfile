# Copyright (c) 2020, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

FROM container-registry.oracle.com/os/oraclelinux:7-slim@sha256:9b86d1332a883ee8f68dd44ba42133de518b2e0ec1cc70257e59fb4da86b1ad3

RUN yum update -y && yum-config-manager --save --setopt=ol7_ociyum_config.skip_if_unavailable=true \
    && yum clean all; rm -rf /var/cache/yum

ENV OPERATOR=/usr/local/bin/verrazzano-helidon-app-operator \
    USER_UID=1001 \
    USER_NAME=operator

# install operator binary
COPY build/_output/bin/verrazzano-helidon-app-operator ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}

# Copy source tree to image
RUN mkdir -p /root/go/src/github.com/verrazzano/verrazzano-helidon-app-operator
COPY . /root/go/src/github.com/verrazzano/verrazzano-helidon-app-operator

