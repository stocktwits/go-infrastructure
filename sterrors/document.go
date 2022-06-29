package sterrors

import (
	"fmt"
	"io"
	"sort"
)

func GetDocumentMd(w io.Writer, config ErrorConfig, appname string) error {
	_, err := fmt.Fprintf(w, "# Application Errors Summary\n\n")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, "The following table summarize all the errors that can be expect from %s.\n\n", appname)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, "|Error Code|Type|Message|HTTP Code|\n")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, "|:----------|:----------|:----------|:----------|\n")
	if err != nil {
		return err
	}

	codes := []ErrorCode{}
	for code := range config {
		codes = append(codes, code)
	}

	sort.Slice(codes, func(i, j int) bool {
		return codes[i] < codes[j]
	})

	for _, code := range codes {
		info := config[code]
		_, err = fmt.Fprintf(w, "|%d|%s|%s|%d|\n", code, info.Type, info.Message, info.Http_code)
		if err != nil {
			return err
		}

	}

	return nil
}
