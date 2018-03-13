From alpine:3.6

ADD http://147.75.94.115/ros-wait-for /usr/bin/
RUN chmod 755 /usr/bin/ros-wait-for
ADD entry.sh /usr/bin/

ENTRYPOINT ["/usr/bin/entry.sh"]
