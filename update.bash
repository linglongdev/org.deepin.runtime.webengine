#!/bin/bash
set -e
repoURL=http://10.20.64.92:8080/crimson_runtime/stable_20250827
codename=main
sources=(
    qt6-base
    qt6-svg
    qt6-declarative
    qt6-imageformats
    qt6-multimedia
    qt6-speech
    qt6-tools
    qt6-wayland
    qt6-translations
    qt6-5compat
)
sources+=(
    dtk6core
    dtk6declarative
    dtk6gui
    dtk6widget
    dtk6log
    dde-qt6platform-plugins
    qt6integration
    fcitx5-qt
)

# qt webengine
sources+=(
    qt6-webengine
    qt6-webchannel
)

# 解决功能问题
sources+=(
    # 解决 ctrl+shift+? 快捷键对话框
    deepin-shortcut-viewer
)

# 查找源码包的所有二进制包，过滤掉调试符号、例子、文档等非二进制包
rm install.list.tmp || true
for src in "${sources[@]}"; do
    echo "Source $src" >&2
    dir=${src:0:1}
    if [[ $src == lib* ]]; then
        dir=${src:0:4}
    fi
    out=$(curl -q -f "$repoURL/pool/$codename/$dir/$src/" 2>/dev/null | grep deb | awk -F'_' '{print $1}' | awk -F'"' '{print $2}' | uniq)
    echo "  # source package $src" >>install.list.tmp
    for pkg in $(echo "$out" | grep -v 'dbgsym$' | grep -v '\-doc$' | grep -v '\-examples$' | grep -v '\-doc\-'); do
        echo "  Binary $pkg" >&2
        echo "  # linglong:gen_deb_source install $pkg" >>install.list.tmp
    done
done

# 删除依赖qt5的包
sed -i '/libfcitx5-qt1/d' install.list.tmp
sed -i '/libfcitx5-qt-dev/d' install.list.tmp
sed -i '/fcitx5-frontend-qt5/d' install.list.tmp

for file in linglong.yaml arm64/linglong.yaml loong64/linglong.yaml sw64/linglong.yaml mips64/linglong.yaml; do
    # 删除gen_deb_source后面的内容，将二进制包补充进去
    grep -B 1000 'linglong:gen_deb_source sources' $file >$file.new
    cat install.list.tmp >>$file.new
    sed -i "s#id: org.deepin.runtime.dtk\$#id: org.deepin.runtime.webengine#" $file.new
    mv $file.new $file
done

rm install.list.tmp

# for src in "${sources[@]}"; do
#     echo -n '$Source ('$src') | '
# done
# echo 