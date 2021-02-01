package cmdx

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ory/x/stringslice"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	dynamicTable struct {
		t  [][]string
		cs int
	}
	dynamicRow []string
)

var (
	_ Table    = (*dynamicTable)(nil)
	_ TableRow = (*dynamicRow)(nil)
)

func dynamicHeader(l int) []string {
	h := make([]string, l)
	for i := range h {
		h[i] = "C" + strconv.Itoa(i)
	}
	return h
}

func (d *dynamicTable) Header() []string {
	return dynamicHeader(d.cs)
}

func (d *dynamicTable) Table() [][]string {
	return d.t
}

func (d *dynamicTable) Interface() interface{} {
	return d.t
}

func (d *dynamicTable) Len() int {
	return len(d.t)
}

func (d dynamicRow) Header() []string {
	return dynamicHeader(len(d))
}

func (d dynamicRow) Columns() []string {
	return d
}

func (d dynamicRow) Interface() interface{} {
	return d
}

func TestRegisterFlags(t *testing.T) {
	t.Run("case=format flags", func(t *testing.T) {
		t.Run("format=no value", func(t *testing.T) {
			flags := pflag.NewFlagSet("test flags", pflag.ContinueOnError)
			RegisterFormatFlags(flags)

			require.NoError(t, flags.Parse([]string{}))
			f, err := flags.GetString(FlagFormat)
			require.NoError(t, err)

			assert.Equal(t, FormatDefault, format(f))
		})
	})

	t.Run("method=table row", func(t *testing.T) {
		tr := dynamicRow{"0", "1", "2"}
		allFields := append(tr.Header(), tr...)

		for _, tc := range []struct {
			fArgs     []string
			contained []string
		}{
			{
				fArgs:     []string{"--" + FlagFormat, string(FormatTable)},
				contained: allFields,
			},
			{
				fArgs:     []string{"--" + FlagQuiet},
				contained: []string{tr[0]},
			},
			{
				fArgs:     []string{"--" + FlagFormat, string(FormatJSON)},
				contained: tr,
			},
			{
				fArgs:     []string{"--" + FlagFormat, string(FormatJSONPretty)},
				contained: tr,
			},
		} {
			t.Run(fmt.Sprintf("format=%v", tc.fArgs), func(t *testing.T) {
				cmd := &cobra.Command{Use: "x"}
				RegisterFormatFlags(cmd.Flags())

				out := &bytes.Buffer{}
				cmd.SetOut(out)
				require.NoError(t, cmd.Flags().Parse(tc.fArgs))

				PrintRow(cmd, tr)

				for _, s := range tc.contained {
					assert.Contains(t, out.String(), s, "%s", out.String())
				}
				notContained := stringslice.Filter(allFields, func(s string) bool {
					return stringslice.Has(tc.contained, s)
				})
				for _, s := range notContained {
					assert.NotContains(t, out.String(), s, "%s", out.String())
				}

				assert.Equal(t, "\n", out.String()[len(out.String())-1:])
			})
		}
	})

	t.Run("method=table", func(t *testing.T) {
		t.Run("case=full table", func(t *testing.T) {
			tb := &dynamicTable{
				t: [][]string{
					{"a0", "b0", "c0"},
					{"a1", "b1", "c1"},
				},
				cs: 3,
			}
			allFields := append(tb.Header(), append(tb.t[0], tb.t[1]...)...)

			for _, tc := range []struct {
				fArgs     []string
				contained []string
			}{
				{
					fArgs:     []string{"--" + FlagFormat, string(FormatTable)},
					contained: allFields,
				},
				{
					fArgs:     []string{"--" + FlagQuiet},
					contained: []string{tb.t[0][0], tb.t[1][0]},
				},
				{
					fArgs:     []string{"--" + FlagFormat, string(FormatJSON)},
					contained: append(tb.t[0], tb.t[1]...),
				},
				{
					fArgs:     []string{"--" + FlagFormat, string(FormatJSONPretty)},
					contained: append(tb.t[0], tb.t[1]...),
				},
			} {
				t.Run(fmt.Sprintf("format=%v", tc.fArgs), func(t *testing.T) {
					cmd := &cobra.Command{Use: "x"}
					RegisterFormatFlags(cmd.Flags())

					out := &bytes.Buffer{}
					cmd.SetOut(out)
					require.NoError(t, cmd.Flags().Parse(tc.fArgs))

					PrintTable(cmd, tb)

					for _, s := range tc.contained {
						assert.Contains(t, out.String(), s, "%s", out.String())
					}
					notContained := stringslice.Filter(allFields, func(s string) bool {
						return stringslice.Has(tc.contained, s)
					})
					for _, s := range notContained {
						assert.NotContains(t, out.String(), s, "%s", out.String())
					}

					assert.Equal(t, "\n", out.String()[len(out.String())-1:])
				})
			}
		})

		t.Run("case=empty table", func(t *testing.T) {
			tb := &dynamicTable{
				t:  nil,
				cs: 1,
			}

			for _, tc := range []struct {
				fArgs    []string
				expected string
			}{
				{
					fArgs:    []string{"--" + FlagFormat, string(FormatTable)},
					expected: "C0\t",
				},
				{
					fArgs:    []string{"--" + FlagQuiet},
					expected: "",
				},
				{
					fArgs:    []string{"--" + FlagFormat, string(FormatJSON)},
					expected: "null",
				},
				{
					fArgs:    []string{"--" + FlagFormat, string(FormatJSONPretty)},
					expected: "null",
				},
			} {
				t.Run(fmt.Sprintf("format=%v", tc.fArgs), func(t *testing.T) {
					cmd := &cobra.Command{Use: "x"}
					RegisterFormatFlags(cmd.Flags())

					out := &bytes.Buffer{}
					cmd.SetOut(out)
					require.NoError(t, cmd.Flags().Parse(tc.fArgs))

					PrintTable(cmd, tb)

					assert.Equal(t, tc.expected+"\n", out.String())
				})
			}
		})
	})
}