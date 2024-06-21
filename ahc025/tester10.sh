set -xe

make build
(cd ./tools && cargo run -r --bin tester ../bin/a.out < in/0000.txt > out.txt)
(cd ./tools && cargo run -r --bin tester ../bin/a.out < in/0001.txt > out.txt)
(cd ./tools && cargo run -r --bin tester ../bin/a.out < in/0002.txt > out.txt)
(cd ./tools && cargo run -r --bin tester ../bin/a.out < in/0003.txt > out.txt)
(cd ./tools && cargo run -r --bin tester ../bin/a.out < in/0004.txt > out.txt)
(cd ./tools && cargo run -r --bin tester ../bin/a.out < in/0005.txt > out.txt)
(cd ./tools && cargo run -r --bin tester ../bin/a.out < in/0006.txt > out.txt)
(cd ./tools && cargo run -r --bin tester ../bin/a.out < in/0007.txt > out.txt)
(cd ./tools && cargo run -r --bin tester ../bin/a.out < in/0008.txt > out.txt)
(cd ./tools && cargo run -r --bin tester ../bin/a.out < in/0009.txt > out.txt)