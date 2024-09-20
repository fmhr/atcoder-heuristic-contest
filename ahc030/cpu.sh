#/bin/sh
cmd="./a.out -cpuprofile cpu.out"

make build
./tools/target/release/tester $cmd < tools/in/0019.txt > out.txt
go tool pprof -http=localhost:8888 a.out cpu.out