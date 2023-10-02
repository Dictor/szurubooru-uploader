package main

import (
	"fmt"
	"syscall"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	isDebug        bool
	client         = resty.New()
	Logger         = logrus.New()
	ImageExtension = []string{".jpg", ".jpeg", ".png", ".gif", ".webm", ".mp4"}
)

var (
	userToken                       string
	host, userId, userPass          string
	safety                          string
	userCookieName, userCookieValue string
)

func main() {
	var (
		cmdUpload = &cobra.Command{
			Use:              "upload <forder path> <tag>",
			Long:             "upload one folder to szurubooru host\n* forder path: path to target directory\n* tag: tag which will be assigned to images, can use comma seperated multiple tags",
			Short:            "upload one folder to szurubooru host",
			Args:             cobra.MinimumNArgs(1),
			Run:              execUpload,
			PersistentPreRun: credentialInput,
		}
		cmdBatchUpload = &cobra.Command{
			Use:              "bupload <root forder path> <handler>",
			Long:             "batch upload multiple folder to szurubooru host\n* forder path: path to root target directory\n*handler: pixiv=%artist (%user_number), name=%artist, split=%tag1 %tag2...",
			Short:            "batch upload multiple folder to szurubooru host",
			Args:             cobra.ExactArgs(2),
			Run:              execBatchUpload,
			PersistentPreRun: credentialInput,
		}
		cmdDelete = &cobra.Command{
			Use:              "delete <tag> <except favorite>",
			Long:             "delete posts which have given tag\n* tag: tag for deleting\n* except favorite: true=delete except favorited post, false = delete all posts",
			Short:            "delete posts which have given tag",
			Args:             cobra.ExactArgs(2),
			Run:              execDelete,
			PersistentPreRun: credentialInput,
		}
		cmdSanitize = &cobra.Command{
			Use:              "sanitize <implication>",
			Long:             "delete posts which have belong on tag without favorited which has specific implication tag\n* implication: target implication tag",
			Short:            "sanitize posts belong in appointed condition",
			Args:             cobra.ExactArgs(1),
			Run:              execBatchDelete,
			PersistentPreRun: credentialInput,
		}
	)

	Logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		DisableColors: false,
		ForceQuote:    false,
	})

	var rootCmd = &cobra.Command{Use: "app",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {

		},
	}

	cobra.OnInitialize(func() {
		if isDebug {
			Logger.Infof("Debug logging is enabled!")
			Logger.SetLevel(logrus.DebugLevel)
		}
	})

	rootCmd.AddCommand(cmdUpload, cmdBatchUpload, cmdDelete, cmdSanitize)
	rootCmd.PersistentFlags().StringVar(&host, "host", "http://localhost", "address of host")
	rootCmd.PersistentFlags().StringVar(&safety, "safety", "unsafe", "safety of images in directory")
	rootCmd.PersistentFlags().StringVar(&userId, "uid", "", "user's login id")
	rootCmd.PersistentFlags().StringVar(&userPass, "upw", "", "user's login password")
	rootCmd.PersistentFlags().BoolVar(&isDebug, "debug", false, "print debug log")
	rootCmd.PersistentFlags().StringVar(&userCookieName, "ckname", "", "user cookie name (when set, ./cookie.txt is read and used as value. Disabled if blank)")
	rootCmd.Execute()
}

func credentialInput(cmd *cobra.Command, args []string) {
	var err error
	Logger.Infof("using szurubooru host as '%s'\n", host)
	if len(userId) < 1 {
		fmt.Print("enter user id : ")
		fmt.Scanln(&userId)
	}
	if len(userPass) < 1 {
		fmt.Print("enter user password : ")
		if bytepw, err := terminal.ReadPassword(int(syscall.Stdin)); err != nil {
			Logger.WithError(err).Fatalln("fail to read password input")
		} else {
			userPass = string(bytepw)
			fmt.Println("")
		}
	}
	if userToken, err = login(host, userId, userPass); err != nil {
		Logger.WithError(err).Fatalln("fail to log in")
	}
	Logger.Infoln("login successfully")
}
