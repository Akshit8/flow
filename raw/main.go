package main

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type result struct {
	path string
	sum  [md5.Size]byte
	err  error
}

func main() {
	data, err := md5All(os.Args[1])
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	for name, sum := range data {
		fmt.Printf("%s :: %x\n", name, sum)
	}
}

func md5All(root string) (map[string][md5.Size]byte, error) {
	quit := make(chan struct{})
	defer close(quit)

	c, errc := sumFiles(quit, root)
	if err := <-errc; err != nil {
		return nil, err
	}

	m := make(map[string][md5.Size]byte)
	for res := range c {
		if res.err != nil {
			return nil, res.err
		}
		m[res.path] = res.sum
	}

	return m, nil
}

func sumFiles(quit <-chan struct{}, root string) (<-chan result, <-chan error) {
	c := make(chan result)
	errc := make(chan error, 1)

	go func() {
		var wg sync.WaitGroup

		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.Mode().IsRegular() {
				return nil
			}

			wg.Add(1)
			go func() {
				defer wg.Done()

				data, err := ioutil.ReadFile(path)
				select {
				case c <- result{path, md5.Sum(data), err}:
				case <-quit:
				}
			}()

			select {
			case <-quit:
				return errors.New("walk canceled")
			default:
				return nil
			}
		})

		go func() {
			wg.Wait()
			close(c)
		}()

		errc <- err
	}()

	return c, errc
}
