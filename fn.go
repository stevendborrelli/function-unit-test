package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/logging"

	fnv1beta1 "github.com/crossplane/function-sdk-go/proto/v1beta1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/response"

	"github.com/stevendborrelli/function-unit-test/input/v1beta1"
)

type TestResult struct {
	Description string `json:"description"`
	Assertion   string `json:"assert"`
	Error       string `json:"error,omitempty"`
}

type TestResults struct {
	Passing []TestResult `json:"passing"`
	Failing []TestResult `json:"failing"`
	Error   []TestResult `json:"errors,omitempty"`
}

type Function struct {
	fnv1beta1.UnimplementedFunctionRunnerServiceServer

	log logging.Logger
}

// RunFunction runs the Function.
func (f *Function) RunFunction(_ context.Context, req *fnv1beta1.RunFunctionRequest) (*fnv1beta1.RunFunctionResponse, error) {
	f.log.Info("Running function", "tag", req.GetMeta().GetTag())

	rsp := response.To(req, response.DefaultTTL)

	in := &v1beta1.Input{}
	if err := request.GetInput(req, in); err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get Function input from %T", req))
		return rsp, nil
	}

	if len(in.TestCases) == 0 {
		f.log.Info("No test cases supplied")
		return rsp, nil
	}

	tr := TestResults{}

	// Set up input variables for our tests
	vars := RunFunctionRequestToCELVars(req)

	for i, tc := range in.TestCases {
		f.log.Info(fmt.Sprintf("running test case %d: %s", i, tc.Description))

		res, err := Assert(tc.Assert, vars)
		if err != nil {
			tr.Error = append(tr.Error, TestResult{
				Description: tc.Description,
				Assertion:   tc.Assert,
				Error:       err.Error()})
			continue
		}
		if res {
			tr.Passing = append(tr.Passing, TestResult{
				Description: tc.Description,
				Assertion:   tc.Assert,
			})
		} else {
			tr.Failing = append(tr.Failing, TestResult{
				Description: tc.Description,
				Assertion:   tc.Assert,
			})
		}
	}

	response.Normalf(rsp, "I was run with input %q!", in.TestCases)

	trJson, _ := json.Marshal(tr)

	f.log.Info(string(trJson))

	if in.ErrorOnFailedTest && len(tr.Failing) > 0 {
		tf, _ := json.MarshalIndent(tr.Failing, "", "  ")
		response.Fatal(rsp, errors.Errorf("failing tests: %s", tf))
		return rsp, errors.Errorf("failing tests: %s", tf)
	}

	if len(tr.Error) > 0 {
		te, _ := json.MarshalIndent(tr.Error, "", "  ")
		response.Fatal(rsp, errors.Errorf("tests with errors: %s", te))
		return rsp, errors.Errorf("there were tests with errors: %s", string(te))
	}

	return rsp, nil
}
