FROM golang:1.9.2-stretch

# Install clang
RUN apt-get update -q && \
    apt-get install -y -q clang-3.9 clang-tidy-3.9 clang-format-3.9 && \
    ln -s /usr/bin/clang-3.9 /usr/local/bin/clang && \
    ln -s /usr/bin/clang-tidy-3.9 /usr/local/bin/clang-tidy && \
    ln -s /usr/bin/clang-format-3.9 /usr/local/bin/clang-format && \
    rm -rf /var/lib/apt/lists/*

