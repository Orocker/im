package main

func main() {
	server := NewServer("0.0.0.0", 8888)
	server.Start()
}
