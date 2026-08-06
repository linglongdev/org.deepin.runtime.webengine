# 安装依赖
bash ./install_dep linglong/sources "$PREFIX"
# 添加qmake配置
mkdir -p "$PREFIX/usr/lib/qt6/bin"
envsubst <qt6.conf >"$PREFIX/usr/lib/qt6/bin/qt6.conf"
# 设置环境变量
mkdir -p $PREFIX/etc
cp profile $PREFIX/etc/
# 删除字体文件
rm -rf "$PREFIX/share/fonts"
/runtime/usr/lib/qt6/bin/qmake6 --version
