package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/vo2021/ossync/jsondiff"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

var (
	profile  = "profile-name"
	bucket   = "bucket-name"
	output   = "."
	interval = 10 // seconds
	debug    = false
)

// IsFile Verifies if the path is valid
func IsFile(filePath string) bool {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// IsDir Verifies if the directory is valid
func IsDir(filePath string) bool {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// CreateFile create a file
func CreateFile(fileFolder, fileName string, delete bool) error {
	filePath := path.Join(fileFolder, fileName)
	if IsDir(filePath) {
		return fmt.Errorf("the %s is a dir, cannot creat create file", filePath)
	}

	if IsFile(filePath) {
		if !delete {
			return nil
		}
		if err := os.Remove(filePath); err != nil {
			return err
		}
	}
	err := os.MkdirAll(fileFolder, os.ModePerm)
	if err != nil {
		return err
	}
	_, err = os.Create(filePath)
	return err
}

func writeFile(filename, contents string) {
	file, err := os.Create(filename)
	if err != nil {
		return
	}
	err = file.Truncate(0)
	_, err = file.Seek(0, 0)
	defer file.Close()
	file.WriteString(contents)
}

func get_config_folder() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	home := usr.HomeDir
	configFolder := path.Join(home, ".config/ossync")
	return configFolder
}

func get_metadata_filename(output, bucket string) string {
	configFolder := get_config_folder()
	output = strings.Trim(output, "/~.")
	if len(output) == 0 {
		output, _ = os.Getwd()
	}
	f, err := createFolder(configFolder)
	if f == nil && err != nil {
		println("create folder failed: " + fmt.Sprintf("error(%v)", err))
	}
	name := configFolder + "/" + strings.Replace(output, "/", "_", -1) + "-" + bucket + ".json"
	if debug {
		fmt.Println("metadata file name: " + name)
	}
	return name
}

func get_local_bucket_metadata(output, bucket string) map[string]interface{} {
	fileName := get_metadata_filename(output, bucket)
	if b, _ := existsFile(fileName); !b {
		return nil //make(map[string]interface{})
	}
	dat, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(bucket + " not found!")
	}

	if debug {
		fmt.Println("metadata: " + string(dat))
	}

	var result map[string]interface{}
	json.Unmarshal(dat, &result)
	return result
}

var syncing = false

func cleanup() {
	fmt.Println("cleanup")
	if syncing {
		fmt.Println("syncing in progress, terminating...")
	}
	for syncing {
		time.Sleep(1)
	}
}

func expandHome(path string) string {
	usr, _ := user.Current()
	dir := usr.HomeDir
	if path == "~" {
		path = dir
	} else if strings.HasPrefix(path, "~/") {
		path = filepath.Join(dir, path[2:])
	}
	return path
}

func main() {
	flag.StringVar(&bucket, "bucket", "bucket-name", "the OCI bucket which is synced to local")
	flag.StringVar(&profile, "profile", "DEFAULT", "the OCI profile name")
	flag.StringVar(&output, "output", "", "the local folder path to sync to")
	flag.IntVar(&interval, "interval", 10, "the interval between sync")
	flag.BoolVar(&debug, "debug", false, "debug mode")
	flag.Parse()

	output = expandHome(output)

	mainLoop()
}

func mainLoop() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanup()
		os.Exit(1)
	}()

	for {
		syncing = true

		// cmd := fmt.Sprintf("oci os object list --all -bn %s --profile %s >  %s.json", bucket, profile, bucket)
		cmd := fmt.Sprintf("oci os object list --all -bn %s --profile %s", bucket, profile)
		if debug {
			fmt.Println(cmd)
		}
		out, err := exec.Command("bash", "-c", cmd).Output()
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				fmt.Println("💀 " + string(exitError.Stderr))
			}
			log.Fatalf("💀 bad cmd: %s", cmd)
		}
		if debug {
			fmt.Println(string(out))
		}

		// https://www.sohamkamani.com/golang/parsing-json/
		var result map[string]interface{}
		json.Unmarshal(out, &result)

		cached := get_local_bucket_metadata(output, bucket)
		if diffs := jsondiff.JSONDiff(cached, result, true, ""); len(diffs) == 0 {
			fmt.Print("🙈")
		} else {
			/*for i, v := range diffs {
				fmt.Printf("%n: %s\n", i, v)
			}*/
			fmt.Print("\n🐵\n")
			var oldData []interface{}
			if cached != nil {
				oldData = cached["data"].([]interface{})
			}
			newData := result["data"].([]interface{})

			o := make(map[string]string)
			for _, row := range oldData {
				t := row.(map[string]interface{})
				k := t["name"].(string)
				o[k] = t["md5"].(string)
			}

			for _, row := range newData {
				t := row.(map[string]interface{})
				name := t["name"].(string)
				if val, ok := o[name]; ok {
					if val == t["md5"] {
						delete(o, name)
						continue
					}
				}
				if strings.HasSuffix(name, "/") {
					continue
				}
				fmt.Printf("sync %s\n", name)
				if strings.ContainsAny(name, "/") {
					i := strings.LastIndex(name, "/")
					if output != "" && !strings.HasSuffix(output, "/") {
						output += "/"
					}
					folder := output + name[:i+1]
					createFolder(folder)
					outfile := output + name
					cmd := fmt.Sprintf("oci os object get --name '%s' --file '%s' -bn %s --profile %s", name, outfile, bucket, profile)
					if debug {
						fmt.Println(cmd)
					}
					_, err := exec.Command("bash", "-c", cmd).Output()
					if err != nil {
						log.Fatal(err)
					}
				}
			}
			writeFile(get_metadata_filename(output, bucket), string(out))
		}
		syncing = false
		time.Sleep(time.Duration(interval) * time.Second)
	}

}

func createFolder(p string) (*os.File, error) {
	if b, _ := existsFile(p); b {
		return nil, nil
	}
	if err := os.MkdirAll(filepath.Dir(p), 0770); err != nil {
		return nil, err
	}
	return os.Create(p)
}

// exists returns whether the given file or directory exists
func existsFile(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
