FROM centos:7

LABEL maintainer="Cody Chen <cody@acom-networks.com>" \
  org.label-schema.name="Auto RPM" \
  org.label-schema.vendor="Cody Chen" \
  org.label-schema.schema-version="1.0"

RUN yum install epel-release -y
RUN yum install -y rpm-build git make

RUN mkdir -p /root/rpmbuild/{SPECS,SOURCES,BUILD,BUILDROOT,RPMS,SRPMS}

ADD release/linux/amd64/auto-rpm /bin/

# ENTRYPOINT ["/bin/auto-rpm"]