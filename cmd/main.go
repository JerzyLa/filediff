package main

import (
	"filediff"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var chunkSize int

var rootCmd = &cobra.Command{
	Use:   "filediff",
	Short: "Rolling hash file diff CLI",
}

var signatureCmd = &cobra.Command{
	Use:   "signature [input filename] [signature filename]",
	Short: "Calculate the signature of the input file",
	Long:  "Calculate the signature of the input file and store it in the destination",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if args[0] == "" {
			logrus.Fatal("Missing the input filename")
		}
		if args[1] == "" {
			logrus.Fatal("Missing the signature filename")
		}

		input, err := os.Open(args[0])
		if err != nil {
			logrus.Fatal(err)
		}
		defer input.Close()

		signature, err := os.OpenFile(args[1], os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(0600))
		if err != nil {
			logrus.Fatal(err)
		}
		defer signature.Close()

		_, err = filediff.Signature(input, signature, uint32(chunkSize))
		if err != nil {
			logrus.Fatal(err)
		}
	},
}

var deltaCmd = &cobra.Command{
	Use:   "delta [signature filename] [newfile filename] [delta filename]",
	Short: "Calculates the delta between files",
	Long:  "Calculates the delta between new and old file using created signature",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		if args[0] == "" {
			logrus.Fatal("Missing the signature filename")
		}
		if args[1] == "" {
			logrus.Fatal("Missing the newfile filename")
		}
		if args[2] == "" {
			logrus.Fatal("Missing the delta filename")
		}

		signature, err := filediff.ReadSignatureFile(args[0])
		if err != nil {
			logrus.Fatal(err)
		}

		newfile, err := os.Open(args[1])
		if err != nil {
			logrus.Fatal(err)
		}
		defer newfile.Close()

		delta, err := os.OpenFile(args[2], os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(0600))
		if err != nil {
			logrus.Fatal(err)
		}
		defer delta.Close()

		err = filediff.Delta(signature, newfile, delta)
		if err != nil {
			logrus.Fatal(err)
		}
	},
}

func main() {
	signatureCmd.Flags().IntVarP(&chunkSize, "chunksize", "s", 512, "chunksize for the signature")
	rootCmd.AddCommand(signatureCmd, deltaCmd)
	err := rootCmd.Execute()
	if err != nil {
		logrus.Fatal(err)
	}
}
