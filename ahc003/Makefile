build: main.go
	go build

toolrun:
	go build
	cd tools &&\
		cargo run --release --bin tester in/0000.txt ../solver > out.txt

toolrunbin:
	go build
	./tools/target/release/tester tools/in/0000.txt ./solver > out.txt

localrun:
	go build
	./solver -local < tools/in/0000.txt > out.txt

scriptrun:
	go build
	cd script && go build
	./script/script

toolsbuild:
	cargo build --manifest-path=tools/Cargo.toml --release

 
# vis:
# 	cd tools &&\
# 	cargo run --release --bin vis example.in example.out
# 
# TODO リアクティブ問題でどうやってうごかすか 
# pprof:
# 	./solver -cpuprofile cpu.prof < in/0000.txt
# 	pprof -http=localhost:8080 cpu.prof
