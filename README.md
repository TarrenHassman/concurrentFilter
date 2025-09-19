# Command Structure

	filter <filename> <regex> <output>
	filter <filename> <regex> <output> --select
		--select will create a new file with the selected text
	filter <filename> <regex> <output> --replace <replace>
		--replace will replace the text
	filter <directory> <regex> <output> --recursive
		--recursive recursive on all files in directory
	filter <directory> <regex> <output> --directory
		--directory process all files in a directory
