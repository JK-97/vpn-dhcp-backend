FROM scratch

LABEL maintainer="lrh3321<liuruihua@jiangxing.ai>"

COPY dhcp-backend /dhcp-backend

# Expose client and management ports
EXPOSE 1054

# Run with default memory based store 
ENTRYPOINT ["/dhcp-backend"]
