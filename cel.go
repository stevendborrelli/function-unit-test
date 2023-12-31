package main

import (
	"reflect"

	"github.com/crossplane/crossplane-runtime/pkg/errors"
	fnv1beta1 "github.com/crossplane/function-sdk-go/proto/v1beta1"
	"github.com/google/cel-go/cel"
)

// NewCELEnvironment sets up the CEL Environment
func NewCELEnvironment() (*cel.Env, error) {
	return cel.NewEnv(
		cel.Types(&fnv1beta1.State{}),
		cel.Variable("observed", cel.ObjectType("apiextensions.fn.proto.v1beta1.State")),
		cel.Variable("desired", cel.ObjectType("apiextensions.fn.proto.v1beta1.State")),
	)
}

// ToCELVars formats a RunFunctionRequest for CEL evaluation
func RunFunctionRequestToCELVars(req *fnv1beta1.RunFunctionRequest) map[string]any {
	vars := make(map[string]any)
	vars["desired"] = req.GetDesired()
	vars["observed"] = req.GetObserved()
	return vars
}

// EvaluateCondition will evaluate a CEL expression
func Assert(assertion string, vars map[string]any) (bool, error) {
	if assertion == "" {
		return false, errors.New("Empty Assertion")
	}

	env, err := NewCELEnvironment()
	if err != nil {
		return false, errors.Wrap(err, "CEL Env error")
	}

	ast, iss := env.Parse(assertion)
	if iss.Err() != nil {
		return false, errors.Wrap(iss.Err(), "CEL Parse error")
	}

	// Type-check the expression for correctness.
	checked, iss := env.Check(ast)
	// Report semantic errors, if present.
	if iss.Err() != nil {
		return false, errors.Wrap(iss.Err(), "CEL TypeCheck error")
	}

	// Ensure the output type is a bool.
	if !reflect.DeepEqual(checked.OutputType(), cel.BoolType) {
		return false, errors.Errorf(
			"CEL Type error: expression '%s' must return a boolean, got %v instead",
			assertion,
			checked.OutputType())
	}

	// Plan the program.
	program, err := env.Program(checked)
	if err != nil {
		return false, errors.Wrap(err, "CEL program plan")
	}

	// Evaluate the program without any additional arguments.
	result, _, err := program.Eval(vars)
	if err != nil {
		return false, errors.Wrap(err, "CEL program Evaluation")
	}

	ret, ok := result.Value().(bool)
	if !ok {
		return false, errors.Wrap(err, "CEL program did not return a bool")
	}

	return bool(ret), nil
}
