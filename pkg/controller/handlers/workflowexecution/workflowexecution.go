package workflowexecution

import (
	"context"
	"time"

	"github.com/obot-platform/nah/pkg/apply"
	"github.com/obot-platform/nah/pkg/router"
	"github.com/obot-platform/obot/apiclient/types"
	"github.com/obot-platform/obot/pkg/controller/handlers/workflowstep"
	"github.com/obot-platform/obot/pkg/invoke"
	v1 "github.com/obot-platform/obot/pkg/storage/apis/obot.obot.ai/v1"
	"github.com/obot-platform/obot/pkg/system"
	apierror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Handler struct {
	invoker *invoke.Invoker
}

func New(invoker *invoke.Invoker) *Handler {
	return &Handler{
		invoker: invoker,
	}
}

func (h *Handler) UpdateRun(req router.Request, _ router.Response) error {
	we := req.Object.(*v1.WorkflowExecution)
	if we.Status.State != types.WorkflowStateComplete && we.Status.State != types.WorkflowStateError || we.Spec.RunName == "" {
		return nil
	}

	var run v1.Run
	if err := req.Get(&run, we.Namespace, we.Spec.RunName); apierror.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	for _, result := range run.Spec.ExternalCallResults {
		if result.ID == we.Name {
			// Already set
			return nil
		}
	}

	if run.Status.ExternalCall != nil && run.Status.ExternalCall.ID == we.Name {
		data := we.Status.Output
		if we.Status.Error != "" {
			data = we.Status.Error
		}
		run.Spec.ExternalCallResults = append(run.Spec.ExternalCallResults, v1.ExternalCallResult{
			ID:   we.Name,
			Data: data,
		})
		return req.Client.Update(req.Ctx, &run)
	}

	return nil
}

func (h *Handler) Run(req router.Request, _ router.Response) error {
	var (
		we = req.Object.(*v1.WorkflowExecution)
	)

	if we.Status.State.IsTerminal() {
		if we.Spec.WorkflowGeneration != we.Status.WorkflowGeneration {
			we.Status.State = types.WorkflowStatePending
			we.Status.EndTime = nil
		}
		return nil
	}

	var wf v1.Workflow
	if err := req.Get(&wf, we.Namespace, we.Spec.WorkflowName); err != nil {
		return err
	}

	if err := h.loadManifest(req, we); err != nil {
		return err
	}

	if we.Status.ThreadName == "" {
		t, err := h.newThread(req.Ctx, req.Client, &wf, we)
		if err != nil {
			return err
		}

		we.Status.ThreadName = t.Name
		if err = req.Client.Status().Update(req.Ctx, we); err != nil {
			return err
		}
	} else {
		var thread v1.Thread
		if err := req.Get(&thread, we.Namespace, we.Status.ThreadName); err != nil {
			return err
		}

		var (
			found = false
			input = normalizeInput(we.Spec.Input)
		)
		for i, env := range thread.Spec.Env {
			if env.Name == "WORKFLOW_INPUT" {
				found = true
				if env.Value != input {
					thread.Spec.Env[i].Value = input
					if err := req.Client.Update(req.Ctx, &thread); err != nil {
						return err
					}
				}
			}
		}

		if !found {
			thread.Spec.Env = append(thread.Spec.Env, types.EnvVar{
				Name:  "WORKFLOW_INPUT",
				Value: input,
			})
			if err := req.Client.Update(req.Ctx, &thread); err != nil {
				return err
			}
		}
	}

	var (
		steps        []kclient.Object
		lastStepName string
	)

	for _, step := range we.Status.WorkflowManifest.Steps {
		newStep := workflowstep.NewStep(we.Namespace, we.Name, lastStepName, we.Spec.WorkflowGeneration, step)
		steps = append(steps, newStep)
		lastStepName = newStep.Name
	}

	if we.Status.WorkflowManifest.Output != "" {
		newStep := workflowstep.NewStep(we.Namespace, we.Name, lastStepName, we.Spec.WorkflowGeneration, types.Step{
			ID:   "output",
			Step: we.Status.WorkflowManifest.Output,
		})
		steps = append(steps, newStep)
	}

	_, output, newState, err := workflowstep.GetStateFromSteps(req.Ctx, req.Client, we.Spec.WorkflowGeneration, steps...)
	if err != nil {
		return err
	}

	if newState.IsBlocked() {
		we.Status.State = newState
		we.Status.Error = output
		return apply.New(req.Client).Apply(req.Ctx, req.Object, steps...)
	}

	switch newState {
	case types.WorkflowStateComplete:
		we.Status.Output = output
	case types.WorkflowStateError:
		we.Status.Error = output
	}

	we.Status.State = newState
	we.Status.WorkflowGeneration = we.Spec.WorkflowGeneration
	if we.Status.State.IsTerminal() && we.Status.EndTime == nil {
		we.Status.EndTime = &metav1.Time{Time: time.Now()}
	}

	return apply.New(req.Client).Apply(req.Ctx, req.Object, steps...)
}

func (h *Handler) loadManifest(req router.Request, we *v1.WorkflowExecution) error {
	var wf v1.Workflow
	if err := req.Get(&wf, we.Namespace, we.Spec.WorkflowName); err != nil {
		return err
	}

	we.Status.WorkflowManifest = &wf.Spec.Manifest
	return nil
}

func normalizeInput(input string) string {
	if input == "" {
		return "{}"
	}
	return input
}

func (h *Handler) newThread(ctx context.Context, c kclient.Client, wf *v1.Workflow, we *v1.WorkflowExecution) (*v1.Thread, error) {
	var projectThread v1.Thread
	if err := c.Get(ctx, router.Key(wf.Namespace, wf.Spec.ThreadName), &projectThread); err != nil {
		return nil, err
	}

	thread := v1.Thread{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    wf.Namespace,
			GenerateName: system.ThreadPrefix,
			Finalizers:   []string{v1.ThreadFinalizer},
		},
		Spec: v1.ThreadSpec{
			ParentThreadName:      projectThread.Name,
			AgentName:             projectThread.Spec.AgentName,
			WorkflowName:          we.Spec.WorkflowName,
			WorkflowExecutionName: we.Name,
			WebhookName:           we.Spec.WebhookName,
			EmailReceiverName:     we.Spec.EmailReceiverName,
			CronJobName:           we.Spec.CronJobName,
			UserID:                projectThread.Spec.UserID,
			Env: []types.EnvVar{
				{
					Name:  "WORKFLOW_INPUT",
					Value: normalizeInput(we.Spec.Input),
				},
			},
			SystemTools: []string{system.WorkflowTool, system.TasksWorkflowTool},
		},
	}

	return &thread, c.Create(ctx, &thread)
}

func (h *Handler) ReassignThread(req router.Request, _ router.Response) error {
	var (
		wfe = req.Object.(*v1.WorkflowExecution)
	)

	if wfe.Status.ThreadName != "" || wfe.Spec.WorkflowName == "" {
		return nil
	}

	var we v1.Workflow
	if err := req.Get(&we, wfe.Namespace, wfe.Spec.WorkflowName); err != nil {
		return kclient.IgnoreNotFound(err)
	}

	if we.Spec.ThreadName != "" {
		wfe.Spec.ThreadName = we.Spec.ThreadName
		return req.Client.Update(req.Ctx, wfe)
	}

	return nil
}
