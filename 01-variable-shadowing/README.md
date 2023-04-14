## 01. Variable Shadowing

`main.go` creates a simple CLI tool with simple usage. To build it run:
```bash
go build -o app main.go
```

Then run the followig to see the usage:
```bash
./app -h
Usage: app [OPTIONS]

Options:
  -o <file>     Write the output to a file.
  -v            Enable verbose logging.
  -q            Disable all logging. Useful in CI/CD pipelines.
  -h            Show this help message.


```

### Instructions

1. Explore `main.go` and try to figure out why the program doesn't work.
2. Try to execute the program and see what happens. 
3. Try to run `./app -o output.csv` and see what happens.
4. Fix the problem and make the program work.
   5. to test run `go build -o app main.go` and then `./app -o output.csv` and see if the output file is created.
6. Check out the solution branch with `git checkout 01-variable-shadowing-solution`
