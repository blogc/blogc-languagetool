package main

import (
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	defaultApiUrl = "https://languagetool.org/api"

	cfgFile string

	cmd = &cobra.Command{
		Use:   "blogc-languagetool [SOURCE]",
		Short: "Check grammar of blogc source files using LanguageTool API",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			htmlStr, err := blogcParse(args[0])
			if err != nil {
				logrus.Fatal(err)
			}

			textStr, err := html2text(htmlStr)
			if err != nil {
				logrus.Fatal(err)
			}

			if err := ltCheck(
				viper.GetString("api-url"),
				textStr,
				viper.GetString("language"),
				viper.GetString("mother-tongue"),
				viper.GetStringSlice("disable-rules"),
				viper.GetStringSlice("disable-categories"),
			); err != nil {
				logrus.Fatal(err)
			}
		},
	}
)

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

		viper.AutomaticEnv()

		if err := viper.ReadInConfig(); err == nil {
			logrus.WithField("config", viper.ConfigFileUsed()).Info("Using config file")
		}
	})

	cmd.Flags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.blogc-languagetool.yaml)")
	cmd.Flags().StringP("api-url", "a", defaultApiUrl, fmt.Sprintf("LanguageTool API URL (default is %s)", defaultApiUrl))
	cmd.Flags().StringP("language", "l", "en-US", "source language, e.g. en-US")
	cmd.Flags().StringP("mother-tongue", "m", "", "mother tongue language, e.g. en-US")
	cmd.Flags().StringSliceP("disable-rule", "r", []string{}, "disable grammar checking rule")
	cmd.Flags().StringSliceP("disable-category", "c", []string{}, "disable grammar checking category")

	viper.BindPFlag("api-url", cmd.Flags().Lookup("api-url"))
	viper.BindPFlag("language", cmd.Flags().Lookup("language"))
	viper.BindPFlag("mother-tongue", cmd.Flags().Lookup("mother-tongue"))
	viper.BindPFlag("disable-rules", cmd.Flags().Lookup("disable-rule"))
	viper.BindPFlag("disable-categories", cmd.Flags().Lookup("disable-category"))
}

func main() {
	cmd.Execute()
}
