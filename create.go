package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/holgerBerger/pfa/pfalib"
)

// create outfile file
func create(args []string) {
	// we scan all files beforehand, to get an idea how big the tree is
	// this wastes some time for large trees, but we scan fast...
	scanner := NewScanner()
	for _, dir := range args {
		scanner.AddDir(dir)
	}

	// start the scanner, this blocks until scanning is done
	scanstart := time.Now()
	scanner.StartScan(opts.Scanners)
	fmt.Printf("scanned %d files in %1.1f seconds, %1.2f files/s.\n",
		len(scanner.Files),
		time.Since(scanstart).Seconds(),
		float64(len(scanner.Files))/time.Since(scanstart).Seconds(),
	)

	// determine compression method
	var compressionmethod pfalib.CompressionType

	compressionmethod = pfalib.NoneC

	if opts.Compression == "snappy" {
		compressionmethod = pfalib.SnappyC
	} else if opts.Compression == "zstd" {
		compressionmethod = pfalib.ZstandardC
	} else if opts.Compression == "none" {
		compressionmethod = pfalib.NoneC
	} else {
		fmt.Fprintln(os.Stderr, "unknown compression method, not compressing.")
	}

	// create outfile
	outfile, err := os.Create(opts.Output)
	if err != nil {
		panic("could not open outfile!")
	}
	boutfile := bufio.NewWriterSize(outfile, int(opts.Blocksize*1024))

	// create archive writer
	archiver := pfalib.NewArchiveWriter(boutfile, opts.Blocksize*1024, opts.Readers, compressionmethod)

	// append all files
	for _, f := range scanner.Files {
		// fmt.Println("adding", f.Path, f.File.Name())
		if f.Path != "" {
			archiver.AppendFile(f)
		} else {
			// ignore markers
		}
	}

	// finalize archive
	files, timediff, bytes, cbytes := archiver.Close()
	boutfile.Flush()
	outfile.Close()

	// print statistics
	fmt.Printf("written %d files in %1.1f seconds with %1.2f MB/s.\n",
		files, timediff.Seconds(), (float64(cbytes)/(timediff.Seconds()))/(1024*1024))
	if compressionmethod != pfalib.NoneC {
		fmt.Printf("%f%% compression.\n", float64(cbytes)/float64(bytes)*100.0)
	}

}

// create outfile file
func createMultiple(args []string, n int) {
	// we scan all files beforehand, to get an idea how big the tree is
	// this wastes some time for large trees, but we scan fast...
	// BUG this version is buggy, use version 2
	scanner := NewScanner()
	for _, dir := range args {
		scanner.AddDir(dir)
	}

	// start the scanner, this blocks until scanning is done
	scanstart := time.Now()
	scanner.StartScan(opts.Scanners)
	fmt.Printf("scanned %d files in %1.1f seconds, %1.2f files/s.\n",
		len(scanner.Files),
		time.Since(scanstart).Seconds(),
		float64(len(scanner.Files))/time.Since(scanstart).Seconds(),
	)

	// determine compression method
	var compressionmethod pfalib.CompressionType

	compressionmethod = pfalib.NoneC

	if opts.Compression == "snappy" {
		compressionmethod = pfalib.SnappyC
	}
	if opts.Compression == "zstd" {
		compressionmethod = pfalib.ZstandardC
	}

	outfile := make([]*os.File, n, n)
	boutfile := make([]*bufio.Writer, n, n)
	archiver := make([]*pfalib.ArchiveWriter, n, n)

	// create outfiles
	for i := 0; i < n; i++ {
		var err error
		outfile[i], err = os.Create(fmt.Sprintf("%s.%d", opts.Output, i))
		if err != nil {
			panic("could not open outfile!")
		}
		boutfile[i] = bufio.NewWriterSize(outfile[i], int(opts.Blocksize*1024))

		// create archive writer
		archiver[i] = pfalib.NewArchiveWriter(boutfile[i], opts.Blocksize*1024, opts.Readers, compressionmethod)
	}

	// simple load balancer
	balancer := make(chan pfalib.DirEntry, 1)
	var balancergroup sync.WaitGroup

	for i := 0; i < n; i++ {
		go func(n int) {
			balancergroup.Add(1)
			for f := range balancer {
				if f.Path != "" {
					archiver[n].AppendFile(f)
				}
			}
			files, timediff, bytes, cbytes := archiver[n].Close()
			boutfile[n].Flush()
			outfile[n].Close()

			// print statistics
			fmt.Printf("written %d files in %1.1f seconds with %1.2f MB/s.\n",
				files, timediff.Seconds(), (float64(cbytes)/(timediff.Seconds()))/(1024*1024))
			if compressionmethod != pfalib.NoneC {
				fmt.Printf("%f%% compression.\n", float64(cbytes)/float64(bytes)*100.0)
			}
			balancergroup.Done()
		}(i)
	}
	for _, f := range scanner.Files {
		balancer <- f
	}
	close(balancer)
	balancergroup.Wait()
}

// create outfile file
// work stealing experiment
func createMultiple2(args []string, n int) {
	// we scan all files beforehand, to get an idea how big the tree is
	// this wastes some time for large trees, but we scan fast...
	scanner := NewScanner()
	for _, dir := range args {
		scanner.AddDir(dir)
	}

	// start the scanner, this blocks until scanning is done
	scanstart := time.Now()
	scanner.StartScan(opts.Scanners)
	fmt.Printf("scanned %d files in %1.1f seconds, %1.2f files/s.\n",
		len(scanner.Files),
		time.Since(scanstart).Seconds(),
		float64(len(scanner.Files))/time.Since(scanstart).Seconds(),
	)

	// determine compression method
	var compressionmethod pfalib.CompressionType

	compressionmethod = pfalib.NoneC

	if opts.Compression == "snappy" {
		compressionmethod = pfalib.SnappyC
	}
	if opts.Compression == "zstd" {
		compressionmethod = pfalib.ZstandardC
	}

	outfile := make([]*os.File, n, n)
	boutfile := make([]*bufio.Writer, n, n)
	archiver := make([]*pfalib.ArchiveWriter, n, n)

	// create outfiles
	for i := 0; i < n; i++ {
		var err error
		outfile[i], err = os.Create(fmt.Sprintf("%s.%d", opts.Output, i))
		if err != nil {
			panic("could not open outfile!")
		}
		boutfile[i] = bufio.NewWriterSize(outfile[i], int(opts.Blocksize*1024))

		// create archive writer
		archiver[i] = pfalib.NewArchiveWriter(boutfile[i], opts.Blocksize*1024, opts.Readers, compressionmethod)
	}

	// simple load balancer
	var balancergroup sync.WaitGroup
	var mutex sync.Mutex

	balancergroup.Add(1)

	for i := 0; i < n; i++ {
		go func(n int) {
			balancergroup.Add(1)

			// directories first
			for ii := 0; ii < len(scanner.Files); ii++ {
				f := scanner.Files[ii]
				if f.Path != "" && f.File.IsDir() {
					archiver[n].AppendFile(f) // FIXME [n] or [i]
				}
			}

			// fles
			var f pfalib.DirEntry
			for {
				mutex.Lock()
				if len(scanner.Files) == 0 {
					mutex.Unlock()
					break
				}
				f, scanner.Files = scanner.Files[len(scanner.Files)-1], scanner.Files[:len(scanner.Files)-1]
				mutex.Unlock()
				runtime.Gosched() // have to call sched here, so others get some data
				if f.Path != "" && !f.File.IsDir() {
					archiver[n].AppendFile(f) // FIXME [i] ??
				}
			}

			files, timediff, bytes, cbytes := archiver[n].Close()
			boutfile[n].Flush()
			outfile[n].Close()

			// print statistics
			fmt.Printf("written %d files in %1.1f seconds with %1.2f MB/s.\n",
				files, timediff.Seconds(), (float64(cbytes)/(timediff.Seconds()))/(1024*1024))
			if compressionmethod != pfalib.NoneC {
				fmt.Printf("%f%% compression.\n", float64(cbytes)/float64(bytes)*100.0)
			}
			balancergroup.Done()
		}(i)
	}
	runtime.Gosched()
	balancergroup.Done()
	balancergroup.Wait()
}
