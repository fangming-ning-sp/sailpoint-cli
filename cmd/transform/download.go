// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sailpoint-oss/sp-cli/client"
	"github.com/spf13/cobra"
)

func newDownloadCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "download",
		Short:   "Download transforms",
		Long:    "Download transforms to local storage",
		Aliases: []string{"dl"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			endpoint := cmd.Flags().Lookup("transforms-endpoint").Value.String()

			resp, err := client.Get(cmd.Context(), endpoint)
			if err != nil {
				return err
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("non-200 response: %s\nbody: %s", resp.Status, body)
			}

			raw, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			// Since we just want to save the content to files, we don't need
			// to parse individual fields.  Just get the string representation.
			var transforms []map[string]interface{}

			err = json.Unmarshal(raw, &transforms)
			if err != nil {
				return err
			}

			outputpath := cmd.Flags().Lookup("output").Value.String()

			for _, v := range transforms {
				filename := v["name"].(string) + ".json"
				content, _ := json.MarshalIndent(v, "", "    ")

				var err error
				if outputpath != "" {
					_ = os.Mkdir(outputpath, os.ModePerm) // Make sure the output dir exists first
					err = ioutil.WriteFile(filepath.Join(outputpath, filename), content, os.ModePerm)
				} else {
					err = ioutil.WriteFile(filename, content, os.ModePerm)
				}

				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	cmd.Flags().StringP("output", "o", "", "The path to the directory to save the files in (default current working directory).")

	return cmd
}
