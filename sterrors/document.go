package sterrors

import (
	"fmt"
	"io"
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

	for code, info := range config {
		_, err = fmt.Fprintf(w, "|%d|%s|%s|%d|\n", code, info.Type, info.Message, info.Http_code)
		if err != nil {
			return err
		}
	}

	return nil
}
