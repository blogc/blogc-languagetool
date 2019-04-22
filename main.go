package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	defaultApiUrl = "https://languagetool.org/api"

	Version = "git"

	cfgFile            string
	disabledRules      []string
	disabledCategories []string
	ignoreWords        []string

	cmd = &cobra.Command{
		Use:     "blogc-languagetool SOURCE",
		Short:   "Check grammar of blogc source files using LanguageTool API",
		Args:    cobra.ExactArgs(1),
		Version: Version,
		Run: func(cmd *cobra.Command, args []string) {
			htmlStr, err := blogcParse(args[0])
			if err != nil {
				logrus.Fatal(err)
			}

			if viper.GetBool("dump-html") {
				fmt.Println(htmlStr)
				return
			}

			textStr, err := html2text(htmlStr)
			if err != nil {
				logrus.Fatal(err)
			}

			if viper.GetBool("dump-text") {
				fmt.Println(textStr)
				return
			}

			if err := ltCheck(
				viper.GetString("api-url"),
				textStr,
				viper.GetString("language"),
				viper.GetString("mother-tongue"),
				mergeSlices(viper.GetStringSlice("ignore-words"), ignoreWords),
				mergeSlices(viper.GetStringSlice("disable-rules"), disabledRules),
				mergeSlices(viper.GetStringSlice("disable-categories"), disabledCategories),
			); err != nil {
				logrus.Fatal(err)
			}
		},
	}
)

func mergeSlices(sl1 []string, sl2 []string) []string {
	rv := []string{}
	for _, s1 := range sl1 {
		rv = append(rv, strings.TrimSpace(s1))
	}

	for _, s2 := range sl2 {
		v := strings.TrimSpace(s2)
		found := false
		for _, r := range rv {
			if v == r {
				found = true
			}
		}
		if !found {
			rv = append(rv, v)
		}
	}

	return rv
}

func init() {
	cobra.OnInitialize(func() {
		if cfgFile != "" {
			viper.SetConfigFile(cfgFile)
		} else {
			home, err := homedir.Dir()
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			viper.AddConfigPath(home)
			viper.SetConfigName(".blogc-languagetool")
		}

		withConfig := viper.ReadInConfig() == nil

		lvl, err := logrus.ParseLevel(viper.GetString("log-level"))
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		logrus.SetLevel(lvl)

		if withConfig {
			logrus.WithField("config", viper.ConfigFileUsed()).Info("Using config file")
		}
	})

	cmd.Flags().StringP(
		"api-url",
		"a",
		defaultApiUrl,
		"LanguageTool API URL",
	)
	cmd.Flags().StringVar(
		&cfgFile,
		"config",
		"",
		"config file (default ~/.blogc-languagetool.yaml)",
	)
	cmd.Flags().StringSliceVarP(
		&disabledCategories,
		"disable-categories",
		"c",
		nil,
		"comma-separated list of grammar checking categories to disable, merged with config",
	)
	cmd.Flags().StringSliceVarP(
		&disabledRules,
		"disable-rules",
		"r",
		nil,
		"comma-separated list of grammar checking rules to disable, merged with config",
	)
	cmd.Flags().BoolP(
		"dump-html",
		"d",
		false,
		"dump HTML generated by blogc and exit without converting to text and checking grammar",
	)
	cmd.Flags().BoolP(
		"dump-text",
		"t",
		false,
		"dump text generated and exit without checking grammar",
	)
	cmd.Flags().StringSliceVarP(
		&ignoreWords,
		"ignore-words",
		"i",
		nil,
		"comma-separated list of words to ignore when checking grammar, merged with config",
	)
	cmd.Flags().StringP(
		"language",
		"l",
		"en-US",
		"source language",
	)
	cmd.Flags().String(
		"log-level",
		logrus.WarnLevel.String(),
		"log level",
	)
	cmd.Flags().StringP(
		"mother-tongue",
		"m",
		"",
		"mother tongue",
	)

	viper.BindPFlag("api-url", cmd.Flags().Lookup("api-url"))
	viper.BindPFlag("dump-html", cmd.Flags().Lookup("dump-html"))
	viper.BindPFlag("dump-text", cmd.Flags().Lookup("dump-text"))
	viper.BindPFlag("language", cmd.Flags().Lookup("language"))
	viper.BindPFlag("log-level", cmd.Flags().Lookup("log-level"))
	viper.BindPFlag("mother-tongue", cmd.Flags().Lookup("mother-tongue"))
}

func main() {
	cmd.Execute()
}
