package formatters

import (
	"bytes"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"os"
)

type YamlFormatter struct{}

func (F* YamlFormatter) Name() string {
	return "yaml"
}

func (F *YamlFormatter) FileExtensions() []string {
	return []string{".yml"}
}

func (F *YamlFormatter) IsInstalled() bool {
	return true
}

func (F *YamlFormatter) FormatToBuffer(args []string, file string, in io.Reader, out io.Writer) error {
	data, err := ioutil.ReadAll(in)
	if err != nil {
		log.Fatal(err)
	}

	t := yaml.MapSlice{}

	err = yaml.Unmarshal([]byte(data), &t)
	if err != nil {
		return err
	}

	d, err := yaml.Marshal(&t)
	if err != nil {
		return err
	}

	// write formatted yml to output
	_, err = out.Write(d)
	if err != nil {
		return err
	}

	return nil
}

func (F *YamlFormatter) FormatInPlace(args[] string, file string) error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	out, err := os.Create(file)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	return F.FormatToBuffer(args, file, bytes.NewReader(data), out)
}
