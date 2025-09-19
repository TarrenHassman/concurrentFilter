package cmd

import (
	"os"
	"regexp"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "filter",
	Short: "Filter file using any regex command, creates a new file with either the filtered text or the unfiltered text",
	Long: `Filter file using any regex command, creates a new file with either the filtered text or the unfiltered text
	
	`,
	Example: `
	filter <filename> <regex> <output>
	filter <filename> <regex> <output> --select
		--select will create a new file with the selected text
	filter <filename> <regex> <output> --replace <replace>
		--replace will replace the text
	filter <directory> <regex> <output> --recursive
		--recursive recursive on all files in directory
	filter <directory> <regex> <output> --directory
		--directory process all files in a directory
	`,
	Args: cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]
		regex := args[1]
		output := args[2]
		selectFlag, _ := cmd.Flags().GetBool("select")
		replace, _ := cmd.Flags().GetString("replace")
		recursive, _ := cmd.Flags().GetBool("recursive")
		directory, _ := cmd.Flags().GetBool("directory")

		if recursive || directory {
			ProcessDirectory(filename, regex, output, selectFlag, directory, recursive, replace)
		} else {

			ProcessFile(filename, regex, output, selectFlag, directory, recursive, replace)
		}
	},
}

func ProcessFile(filename, regex, output string, selectFlag, cgzFlag, czFlag bool, replace string) {
	re := regexp.MustCompile(`.*\.(zip|gz)$`)

	if re.MatchString(filename) {
		compressedFilter()
	} else {
		filter(filename, regex, output, selectFlag, replace)
	}

}

// TODO: Implement concurrent file processing using goroutines and channels
func filter(filename, regex, output string, selectFlag bool, replace string) {

	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		println("Error opening file:", err.Error())
		return
	}
	defer file.Close()

	// Read the file
	content, err := os.ReadFile(filename)
	if err != nil {
		println("Error reading file:", err.Error())
		return
	}

	// Filter the file
	filteredContent := regexp.MustCompile(regex).ReplaceAll(content, []byte(replace))

	// Write the file
	if !selectFlag {
		err = os.WriteFile(output, filteredContent, 0644)
		if err != nil {
			println("Error writing file:", err.Error())
			return
		}
	}
}

// TODO: Implement concurrent file processing using goroutines and channels
func ProcessDirectory(directory, regex, output string, selectFlag, cgzFlag, czFlag bool, replace string) {
	// Walk through the directory
	dir, err := os.ReadDir(directory)
	if err != nil {
		println("Error reading directory:", err.Error())
		return
	}
	for _, entry := range dir {
		// Check if it's a directory
		if entry.IsDir() {
			// Recursively process subdirectory
			ProcessDirectory(directory+string(os.PathSeparator)+entry.Name(), regex, output, selectFlag, cgzFlag, czFlag, replace)
		}
		// Seperate the file name
		filePath := directory + string(os.PathSeparator) + entry.Name()
		// Process the file
		ProcessFile(filePath, regex, output, selectFlag, cgzFlag, czFlag, replace)
	}
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.Flags().BoolP("select", "s", false, "Create a new file with the selected text")
	rootCmd.Flags().StringP("replace", "r", "", "Replace the text")
	rootCmd.Flags().BoolP("recursive", "R", false, "Recursive on all files in directory")
	rootCmd.Flags().BoolP("directory", "d", false, "Process all files in a directory")
}
