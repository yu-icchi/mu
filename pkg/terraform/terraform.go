package terraform

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfcmt "github.com/suzuki-shunsuke/tfcmt/v4/pkg/terraform"
)

//go:generate mkdir -p mock
//go:generate mockgen -source=terraform.go -package=mock -destination=mock/mock.go

type Terraform interface {
	Setup(ctx context.Context) error
	Version(ctx context.Context) (string, map[string]string, error)
	CompareVersion(ctx context.Context, version string) error
	Init(ctx context.Context, params *InitParams, opts ...Option) (*Output, error)
	SwitchWorkspace(ctx context.Context, workspace string) error
	Plan(ctx context.Context, params *PlanParams, opts ...Option) (*Output, error)
	Apply(ctx context.Context, params *ApplyParams, opts ...Option) (*Output, error)
	ForceUnlock(ctx context.Context, lockID string, opts ...Option) (*ForceUnlockOutput, error)
	Import(ctx context.Context, params *ImportParams, opts ...Option) (*ImportOutput, error)
	StateRm(ctx context.Context, params *StateRmParams, opts ...Option) (*StateRmOutput, error)
	Cleanup(ctx context.Context)
}

type Output struct {
	Result             string
	OutsideTerraform   string
	ChangedResult      string
	Warning            string
	HasAddOrUpdateOnly bool
	HasDestroy         bool
	HasNoChanges       bool
	HasError           bool
	HasParseError      bool
	Error              error
	RawLog             string
}

type ForceUnlockOutput struct {
	Result   string
	HasError bool
	Error    error
}

type ImportOutput struct {
	Result   string
	HasError bool
	Error    error
}

type StateRmOutput struct {
	Result   string
	HasError bool
	Error    error
}

const LatestVersion = "latest"

type InitParams struct {
	BackendConfig     map[string]string
	BackendConfigPath string
}

type PlanParams struct {
	Vars     []string
	VarFiles []string
	Destroy  bool
	Out      string
}

type ApplyParams struct {
	PlanFilePath string
}

type ImportParams struct {
	Address  string
	ID       string
	Vars     []string
	VarFiles []string
}

type StateRmParams struct {
	Address string
	DryRun  bool
}

type options struct {
	stream io.Writer
}

type Option func(o *options)

func WithStream(out io.Writer) Option {
	return func(o *options) {
		o.stream = out
	}
}

type installer interface {
	Install(ctx context.Context) (string, error)
	Remove(ctx context.Context) error
}

type terraform struct {
	version   string
	workDir   string
	execPath  string
	installer installer
	tf        *tfexec.Terraform
}

type Params struct {
	Version  string
	WorkDir  string
	ExecPath string
}

func New(params *Params) Terraform {
	return &terraform{
		version:  strings.ToLower(params.Version),
		workDir:  params.WorkDir,
		execPath: params.ExecPath,
	}
}

func (t *terraform) install(ctx context.Context) (string, error) {
	if t.version == LatestVersion {
		t.installer = &releases.LatestVersion{
			Product: product.Terraform,
		}
	} else {
		t.installer = &releases.ExactVersion{
			Product: product.Terraform,
			Version: version.Must(version.NewVersion(t.version)),
		}
	}
	return t.installer.Install(ctx)
}

func (t *terraform) Setup(ctx context.Context) error {
	var err error
	if t.execPath == "" {
		t.execPath, err = t.install(ctx)
		if err != nil {
			return err
		}
	}
	t.tf, err = tfexec.NewTerraform(t.workDir, t.execPath)
	if err != nil {
		return err
	}
	return nil
}

func (t *terraform) Version(ctx context.Context) (string, map[string]string, error) {
	ver, providerVersions, err := t.tf.Version(ctx, true)
	if err != nil {
		return "", nil, err
	}
	providers := make(map[string]string, len(providerVersions))
	for providerName, providerVersion := range providerVersions {
		providers[providerName] = providerVersion.String()
	}
	return ver.String(), providers, nil
}

func (t *terraform) CompareVersion(ctx context.Context, version string) error {
	if version == "" || t.version == LatestVersion {
		return nil
	}
	ver, _, err := t.Version(ctx)
	if err != nil {
		return err
	}
	if ver != version {
		return errTerraformMissmatchVersion
	}
	return nil
}

func (t *terraform) Init(ctx context.Context, params *InitParams, opts ...Option) (*Output, error) {
	opt := &options{}
	for i := range opts {
		opts[i](opt)
	}

	outBuf := new(strings.Builder)
	errBuf := new(strings.Builder)
	if opt.stream != nil {
		t.tf.SetStdout(io.MultiWriter(outBuf, opt.stream))
		t.tf.SetStderr(io.MultiWriter(errBuf, opt.stream))
	} else {
		t.tf.SetStdout(outBuf)
		t.tf.SetStderr(errBuf)
	}
	initOpts := make([]tfexec.InitOption, 0, 1+len(params.BackendConfig))
	if params.BackendConfigPath != "" {
		initOpts = append(initOpts, tfexec.BackendConfig(params.BackendConfigPath))
	}
	for key, value := range params.BackendConfig {
		initOpts = append(initOpts, tfexec.BackendConfig(fmt.Sprintf("%s=%s", key, value)))
	}
	if params.BackendConfigPath != "" || len(params.BackendConfig) > 0 {
		initOpts = append(initOpts, tfexec.Reconfigure(true))
	}
	parser := tfcmt.NewPlanParser()
	if err := t.tf.Init(ctx, initOpts...); err != nil {
		if errBuf.Len() == 0 {
			return nil, err
		}
		ret := parser.Parse(errBuf.String())
		if ret.HasParseError {
			return nil, err
		}
		return t.toOutput(ret, errBuf.String()), nil
	}
	ret := parser.Parse(outBuf.String())
	return t.toOutput(ret, outBuf.String()), nil
}

func (t *terraform) SwitchWorkspace(ctx context.Context, workspace string) error {
	if workspace == "" || workspace == "default" {
		return nil
	}
	err := t.tf.WorkspaceSelect(ctx, workspace)
	if err != nil {
		return t.tf.WorkspaceNew(ctx, workspace)
	}
	return nil
}

func (t *terraform) Plan(ctx context.Context, params *PlanParams, opts ...Option) (*Output, error) {
	opt := &options{}
	for i := range opts {
		opts[i](opt)
	}

	outBuf := new(strings.Builder)
	errBuf := new(strings.Builder)
	if opt.stream != nil {
		t.tf.SetStdout(io.MultiWriter(outBuf, opt.stream))
		t.tf.SetStderr(io.MultiWriter(errBuf, opt.stream))
	} else {
		t.tf.SetStdout(outBuf)
		t.tf.SetStderr(errBuf)
	}
	planOpts := make([]tfexec.PlanOption, 0, len(params.Vars)+len(params.VarFiles)+2)
	for _, v := range params.Vars {
		planOpts = append(planOpts, tfexec.Var(v))
	}
	for _, v := range params.VarFiles {
		planOpts = append(planOpts, tfexec.VarFile(v))
	}
	if params.Out != "" {
		planOpts = append(planOpts, tfexec.Out(params.Out))
	}
	planOpts = append(planOpts, tfexec.Destroy(params.Destroy))

	parser := tfcmt.NewPlanParser()
	_, err := t.tf.Plan(ctx, planOpts...)
	if err != nil {
		if errBuf.Len() == 0 {
			return nil, err
		}
		ret := parser.Parse(errBuf.String())
		if ret.HasParseError {
			return nil, err
		}
		return t.toOutput(ret, errBuf.String()), nil
	}
	ret := parser.Parse(outBuf.String())
	return t.toOutput(ret, outBuf.String()), nil
}

func (t *terraform) Apply(ctx context.Context, params *ApplyParams, opts ...Option) (*Output, error) {
	opt := &options{}
	for i := range opts {
		opts[i](opt)
	}

	outBuf := new(strings.Builder)
	errBuf := new(strings.Builder)
	if opt.stream != nil {
		t.tf.SetStdout(io.MultiWriter(outBuf, opt.stream))
		t.tf.SetStderr(io.MultiWriter(errBuf, opt.stream))
	} else {
		t.tf.SetStdout(outBuf)
		t.tf.SetStderr(errBuf)
	}
	applyOpts := make([]tfexec.ApplyOption, 0, 1)
	if params.PlanFilePath != "" {
		applyOpts = append(applyOpts, tfexec.DirOrPlan(params.PlanFilePath))
	}

	parser := tfcmt.NewApplyParser()
	err := t.tf.Apply(ctx, applyOpts...)
	if err != nil {
		if errBuf.Len() == 0 {
			return nil, err
		}
		ret := parser.Parse(errBuf.String())
		if ret.HasParseError {
			return nil, err
		}
		return t.toOutput(ret, errBuf.String()), nil
	}
	ret := parser.Parse(outBuf.String())
	return t.toOutput(ret, outBuf.String()), nil
}

func (t *terraform) ForceUnlock(ctx context.Context, lockID string, opts ...Option) (*ForceUnlockOutput, error) {
	opt := &options{}
	for i := range opts {
		opts[i](opt)
	}

	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	if opt.stream != nil {
		t.tf.SetStdout(io.MultiWriter(outBuf, opt.stream))
		t.tf.SetStderr(io.MultiWriter(errBuf, opt.stream))
	} else {
		t.tf.SetStdout(outBuf)
		t.tf.SetStderr(errBuf)
	}

	if err := t.tf.ForceUnlock(ctx, lockID); err != nil {
		if errBuf.Len() == 0 {
			return nil, err
		}
		out := &ForceUnlockOutput{
			Result:   errBuf.String(),
			HasError: true,
			Error:    err,
		}
		return out, nil
	}
	out := &ForceUnlockOutput{
		Result: outBuf.String(),
	}
	return out, nil
}

func (t *terraform) Import(ctx context.Context, params *ImportParams, opts ...Option) (*ImportOutput, error) {
	opt := &options{}
	for i := range opts {
		opts[i](opt)
	}

	outBuf := new(strings.Builder)
	errBuf := new(strings.Builder)
	if opt.stream != nil {
		t.tf.SetStdout(io.MultiWriter(outBuf, opt.stream))
		t.tf.SetStderr(io.MultiWriter(errBuf, opt.stream))
	} else {
		t.tf.SetStdout(outBuf)
		t.tf.SetStderr(errBuf)
	}
	importOpts := make([]tfexec.ImportOption, 0, len(params.Vars)+len(params.VarFiles))
	for _, v := range params.Vars {
		importOpts = append(importOpts, tfexec.Var(v))
	}
	for _, v := range params.VarFiles {
		importOpts = append(importOpts, tfexec.VarFile(v))
	}

	if err := t.tf.Import(ctx, params.Address, params.ID, importOpts...); err != nil {
		if errBuf.Len() == 0 {
			return nil, err
		}
		out := &ImportOutput{
			Result:   errBuf.String(),
			HasError: true,
			Error:    err,
		}
		return out, nil
	}
	out := &ImportOutput{
		Result: outBuf.String(),
	}
	return out, nil
}

func (t *terraform) StateRm(ctx context.Context, params *StateRmParams, opts ...Option) (*StateRmOutput, error) {
	opt := &options{}
	for i := range opts {
		opts[i](opt)
	}

	outBuf := new(strings.Builder)
	errBuf := new(strings.Builder)
	if opt.stream != nil {
		t.tf.SetStdout(io.MultiWriter(outBuf, opt.stream))
		t.tf.SetStderr(io.MultiWriter(errBuf, opt.stream))
	} else {
		t.tf.SetStdout(outBuf)
		t.tf.SetStderr(errBuf)
	}
	stateRmOpts := []tfexec.StateRmCmdOption{
		tfexec.DryRun(params.DryRun),
	}

	if err := t.tf.StateRm(ctx, params.Address, stateRmOpts...); err != nil {
		if errBuf.Len() == 0 {
			return nil, err
		}
		out := &StateRmOutput{
			Result:   errBuf.String(),
			HasError: true,
			Error:    err,
		}
		return out, nil
	}
	out := &StateRmOutput{
		Result: outBuf.String(),
	}
	return out, nil
}

func (t *terraform) Cleanup(ctx context.Context) {
	if t.installer != nil {
		_ = t.installer.Remove(ctx)
	}
}

func (t *terraform) toOutput(ret tfcmt.ParseResult, rawLog string) *Output {
	return &Output{
		Result:             ret.Result,
		OutsideTerraform:   ret.OutsideTerraform,
		ChangedResult:      ret.ChangedResult,
		Warning:            ret.Warning,
		HasAddOrUpdateOnly: ret.HasAddOrUpdateOnly,
		HasDestroy:         ret.HasDestroy,
		HasNoChanges:       ret.HasNoChanges,
		HasError:           ret.HasError,
		HasParseError:      ret.HasParseError,
		Error:              ret.Error,
		RawLog:             rawLog,
	}
}
