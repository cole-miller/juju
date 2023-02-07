#!/bin/bash

set -e

source "$(dirname $0)/../env.sh"

MUSL_VERSION="1.2.3"

MUSL_PLACEMENT=${MUSL_PLACEMENT:-"system"}

if [ "${MUSL_PLACEMENT}" = "local" ]; then
    MUSL_PATH=${PROJECT_DIR}/_deps/musl-${BUILD_ARCH}
    MUSL_BIN_PATH=${MUSL_PATH}/output/bin
else
    MUSL_PATH=/usr/local/musl
    MUSL_BIN_PATH=${MUSL_PATH}/bin
fi

musl_install_system() {
    sudo ./configure || { echo "Failed to configure musl"; exit 1; }
    sudo make install || { echo "Failed to install musl"; exit 1; }

    LOCAL_PATH=${PROJECT_DIR}/_deps/musl-${BUILD_ARCH}/output/bin

    mkdir -p ${LOCAL_PATH} || { echo "Failed to create ${MUSL_BIN_PATH}"; exit 1; }
    sudo ln -s ${MUSL_BIN_PATH}/musl-gcc ${LOCAL_PATH}/musl-gcc || { echo "Failed to link musl-gcc"; exit 1; }

    sudo ln -s /usr/include/${BUILD_MACHINE}-linux-gnu/asm ${MUSL_PATH}/include/asm || { echo "Failed to link ${BUILD_MACHINE}-linux-gnu/asm headers"; exit 1; }
    sudo ln -s /usr/include/asm-generic ${MUSL_PATH}/include/asm-generic || { echo "Failed to link asm-generic headers"; exit 1; }
    sudo ln -s /usr/include/linux ${MUSL_PATH}/include/linux || { echo "Failed to link linux header"; exit 1; } 
}

musl_install_local() {
    ./configure --prefix=${MUSL_PATH} || { echo "Failed to configure musl"; exit 1; }
    make install || { echo "Failed to install musl"; exit 1; }

    mkdir -p ${MUSL_BIN_PATH} || { echo "Failed to create ${MUSL_BIN_PATH}"; exit 1; }
    ln -s ${MUSL_PATH}/bin/musl-gcc ${MUSL_BIN_PATH}/musl-gcc || { echo "Failed to link musl-gcc"; exit 1; }

    cd ${PROJECT_DIR}
    ln -s /usr/include/${BUILD_MACHINE}-linux-gnu/asm ${MUSL_PATH}/include/asm || { echo "Failed to link ${BUILD_MACHINE}-linux-gnu/asm headers"; exit 1; }
    ln -s /usr/include/asm-generic ${MUSL_PATH}/include/asm-generic || { echo "Failed to link asm-generic headers"; exit 1; }
    ln -s /usr/include/linux ${MUSL_PATH}/include/linux || { echo "Failed to link linux header"; exit 1; }
}

musl_install() {
    TMP_DIR=$(mktemp -d)
    wget -q https://musl.libc.org/releases/musl-${MUSL_VERSION}.tar.gz -O - | tar -xzf - -C ${TMP_DIR}
    cd ${TMP_DIR}/musl-${MUSL_VERSION}

    if [ "${MUSL_PLACEMENT}" = "local" ]; then
        echo "Installing local musl"
        musl_install_local
    else
        echo "Installing system musl"
        musl_install_system
    fi
}

musl_install_cross_arch() {
    mkdir -p ${MUSL_PATH} || { exit 1; }
    git clone https://github.com/richfelker/musl-cross-make.git ${MUSL_PATH}
    cd ${MUSL_PATH}

    mkdir -p ${MUSL_PATH}/build

    case "${BUILD_ARCH}" in
        amd64)   echo "TARGET=x86_64-linux-musl" >> config.mak ;;
        arm64)   echo "TARGET=aarch64-linux-musl" >> config.mak ;;
        s390x)   echo "TARGET=s390x-linux-musl" >> config.mak ;;
        ppc64le) echo "TARGET=powerpc64le-linux-musl" >> config.mak ;;
        riscv64) echo "TARGET=riscv64-linux-musl" >> config.mak ;;
        *)
            echo "Unsupported architecture ${BUILD_ARCH}"
            exit 1
            ;;
    esac

    echo "OUTPUT=${MUSL_PATH}/output" >> config.mak
    echo "COMMON_CONFIG += CFLAGS=\"-g0 -Os\" CXXFLAGS=\"-g0 -Os\" LDFLAGS=\"-s\"" >> config.mak

    echo "Building musl-${BUILD_ARCH}"
    make install || { exit 1; }

    echo "Linking musl-${BUILD_ARCH} to musl-gcc"
    cd ${MUSL_PATH}/output/bin

    case "${BUILD_ARCH}" in
        amd64) ln -s x86_64-linux-musl-gcc musl-gcc ;;
        arm64) ln -s aarch64-linux-musl-gcc musl-gcc ;;
        s390x) ln -s s390x-linux-musl-gcc musl-gcc ;;
        ppc64le) ln -s powerpc64le-linux-musl-gcc musl-gcc ;;
        riscv64) ln -s riscv64-linux-musl-gcc musl-gcc ;;
        *)
            echo "Unsupported architecture ${BUILD_ARCH}"
            exit 1
            ;;
    esac
}

install() {
    if [ "${BUILD_ARCH}" = "${CURRENT_ARCH}" ]; then
        musl_install
        exit 0
    fi

    echo "Installing cross-arch musl"
    musl_install_cross_arch
    exit 0
}