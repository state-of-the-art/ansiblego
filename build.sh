#!/bin/sh -e
# Pack and creates multibinary executables

# TODO: Using gz for now due to upx isn't working on macos
# as expected and seems related to the code signature issues.
# GZ is used due to it's speed, compatibility and quite small binary size.
# You can set it to 'raw', 'upx' or 'xz' as well
[ "x${PKG_SUFFIX}" != 'x' ] || PKG_SUFFIX='gz'  # To use general xz archiver

name=ansiblego
suffixes="linux-amd64 linux-arm64 windows-amd64 darwin-amd64 darwin-arm64"

# Running static code checks
./check.sh

# Disabling cgo in order to not link with libc and utilize static linkage binaries
# which will help to not relay on glibc on linux and be truely independend from OS
export CGO_ENABLED=0

for suffix in $suffixes; do
    echo "--> Build binary for ${suffix}"
    # Utilizing a number of optimizations to reduce the exec binary size
    GOOS="$(echo "${suffix}" | cut -d- -f1)" GOARCH="$(echo "${suffix}" | cut -d- -f2)" \
        go build -ldflags="-s -w" -gcflags=all="-l -B" -a -o "${name}.raw.${suffix}" "./cmd/${name}" &
done
wait

if [ "x${PKG_SUFFIX}" != 'xraw' ] ; then
    # Pack all the executables with upx
    for exec_suffix in $suffixes; do
        bin_name="${name}.raw.${exec_suffix}"
        out_name="${name}.${PKG_SUFFIX}.${exec_suffix}"
        # Run the packers only if the results are older than raw binary
        if [ ! -f "${out_name}" -o "${out_name}" -ot "${bin_name}" ]; then
            if [ "x${PKG_SUFFIX}" = 'xupx' ] ; then
                echo "--> UPX pack binary for ${exec_suffix}"
                upx --brute -q -9 -o "${out_name}" "${bin_name}" &
            elif [ "x${PKG_SUFFIX}" = 'xxz' ] ; then
                echo "--> XZ pack binary for ${exec_suffix}"
                xz -z -9e -T0 -c "${bin_name}" > "${out_name}" &
            elif [ "x${PKG_SUFFIX}" = 'xgz' ] ; then
                echo "--> Gzip pack binary for ${exec_suffix}"
                gzip -9 -c "${bin_name}" > "${out_name}" &
            fi
        fi
    done

    wait
fi

# Combine the archs together
for out_suffix in $suffixes; do
    echo "--> Combining binaries for ${out_suffix}"
    out_bin="${name}.out.${out_suffix}"
    [ "x$(echo "${out_suffix}" | cut -d- -f1)" != "xwindows" ] || out_bin="${out_bin}.exe"
    if [ "x${PKG_SUFFIX}" = 'xraw' -o "x${PKG_SUFFIX}" = 'xupx' ] ; then
        # RAW and UPX can be used as is
        cp -a "${name}.${PKG_SUFFIX}.${out_suffix}" "${out_bin}"
    else
        # We can't use xz/gz binary as the host one
        cp -a "${name}.raw.${out_suffix}" "${out_bin}"
    fi

    # Combine with the rest of the archs
    for pack_suffix in $suffixes; do
        [ "x${out_suffix}" != "x${pack_suffix}" ] || continue
        echo "-->   + ${pack_suffix}"
        pack_bin="${name}.${PKG_SUFFIX}.${pack_suffix}"
        echo '' >> "${out_bin}"
        echo "--- EMBEDDED_BINARY ${pack_suffix} ${PKG_SUFFIX} ---" >> "${out_bin}"
        cat "${pack_bin}" >> "${out_bin}"
    done
done

# Combine the sh bundle
echo "--> Combining binaries to shell bundle"
out_bin="${name}.out.sh.bundle"
cp -a unix_bundle.sh.head "${out_bin}"
chmod +x "${out_bin}"
for pack_suffix in $suffixes; do
    echo "-->   + ${pack_suffix}"
    pack_bin="${name}.${PKG_SUFFIX}.${pack_suffix}"
    echo '' >> "${out_bin}"
    echo "--- EMBEDDED_BINARY ${pack_suffix} ${PKG_SUFFIX} ---" >> "${out_bin}"
    cat "${pack_bin}" >> "${out_bin}"
done
