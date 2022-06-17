mkdir -p /tmp/fs-shadow-test
cd /tmp/fs-shadow-test
touch c
mv c /tmp/
rm -rf /tmp/c
touch c
mv c bb
mkdir a
rm -rf a
mkdir a
mv a /tmp/
mv /tmp/a .
mv a dd
rm -rf dd
cd -
