package v3_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/v3"
	"code.cloudfoundry.org/cli/command/v3/v3fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("v3-app Command", func() {
	var (
		cmd             v3.V3AppCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v3fakes.FakeV3AppActor
		binaryName      string
		executeErr      error
		app             string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeV3AppActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		app = "some-app"

		cmd = v3.V3AppCommand{
			AppName: app,

			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(sharedaction.NoTargetedOrganizationError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(command.NoTargetedOrganizationError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			_, checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	Context("when the user is not logged in", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = errors.New("some current user error")
			fakeConfig.CurrentUserReturns(configv3.User{}, expectedErr)
		})

		It("return an error", func() {
			Expect(executeErr).To(Equal(expectedErr))
		})
	})

	Context("when the actor does not return an error", func() {
		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				Name: "some-space",
				GUID: "some-space-guid",
			})
			fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
		})

		Context("when there are no instances of any process in the app", func() {
			BeforeEach(func() {
				summary := v3action.ApplicationSummary{
					Application: v3action.Application{
						Name:  "some-app",
						State: "STARTED",
					},
					CurrentDroplet: v3action.Droplet{
						Stack: "cflinuxfs2",
						Buildpacks: []v3action.Buildpack{
							{
								Name:         "ruby_buildpack",
								DetectOutput: "some-detect-output",
							},
							{
								Name:         "some-buildpack",
								DetectOutput: "",
							},
						},
					},
					Processes: []v3action.Process{
						v3action.Process{
							Type:       "console",
							Instances:  []v3action.Instance{},
							MemoryInMB: 128,
						},
						v3action.Process{
							Type:       "worker",
							MemoryInMB: 64,
							Instances:  []v3action.Instance{},
						},
						v3action.Process{
							Type:       "web",
							MemoryInMB: 32,
							Instances:  []v3action.Instance{},
						},
					},
				}
				fakeActor.GetApplicationSummaryByNameAndSpaceReturns(summary, nil, nil)
			})

			It("says no instances are running", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("There are no running instances of this app."))
			})

			It("does not display the instance table", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).NotTo(Say(`state\s+since\s+cpu\s+memory\s+disk`))
			})
		})

		Context("when all the instances in all processes are down", func() {
			BeforeEach(func() {
				summary := v3action.ApplicationSummary{
					Application: v3action.Application{
						Name:  "some-app",
						State: "STARTED",
					},
					CurrentDroplet: v3action.Droplet{
						Stack: "cflinuxfs2",
						Buildpacks: []v3action.Buildpack{
							{
								Name:         "ruby_buildpack",
								DetectOutput: "some-detect-output",
							},
							{
								Name:         "some-buildpack",
								DetectOutput: "",
							},
						},
					},
					Processes: []v3action.Process{
						v3action.Process{
							Type:       "console",
							Instances:  []v3action.Instance{{State: "DOWN"}},
							MemoryInMB: 128,
						},
						v3action.Process{
							Type:       "worker",
							MemoryInMB: 64,
							Instances:  []v3action.Instance{{State: "DOWN"}},
						},
						v3action.Process{
							Type:       "web",
							MemoryInMB: 32,
							Instances:  []v3action.Instance{{State: "DOWN"}},
						},
					},
				}
				fakeActor.GetApplicationSummaryByNameAndSpaceReturns(summary, nil, nil)
			})

			It("says no instances are running", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("There are no running instances of this app."))
			})

			It("does not display the instance table", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).NotTo(Say(`state\s+since\s+cpu\s+memory\s+disk`))
			})
		})

		Context("when there are running instances of the app", func() {
			BeforeEach(func() {
				summary := v3action.ApplicationSummary{
					Application: v3action.Application{
						Name:  "some-app",
						State: "STARTED",
					},
					CurrentDroplet: v3action.Droplet{
						Stack: "cflinuxfs2",
						Buildpacks: []v3action.Buildpack{
							{
								Name:         "ruby_buildpack",
								DetectOutput: "some-detect-output",
							},
							{
								Name:         "some-buildpack",
								DetectOutput: "",
							},
						},
					},
					Processes: []v3action.Process{
						v3action.Process{
							Type:       "console",
							Instances:  []v3action.Instance{},
							MemoryInMB: 128,
						},
						v3action.Process{
							Type:       "worker",
							MemoryInMB: 64,
							Instances: []v3action.Instance{
								v3action.Instance{
									Index:       0,
									State:       "DOWN",
									MemoryUsage: 4000000,
									DiskUsage:   4000000,
									MemoryQuota: 67108864,
									DiskQuota:   8000000,
									Uptime:      int(time.Now().Sub(time.Unix(1371859200, 0)).Seconds()),
								},
							},
						},
						v3action.Process{
							Type:       "web",
							MemoryInMB: 32,
							Instances: []v3action.Instance{
								v3action.Instance{
									Index:       0,
									State:       "RUNNING",
									MemoryUsage: 1000000,
									DiskUsage:   1000000,
									MemoryQuota: 33554432,
									DiskQuota:   2000000,
									Uptime:      int(time.Now().Sub(time.Unix(267321600, 0)).Seconds()),
								},
								v3action.Instance{
									Index:       1,
									State:       "RUNNING",
									MemoryUsage: 2000000,
									DiskUsage:   2000000,
									MemoryQuota: 33554432,
									DiskQuota:   4000000,
									Uptime:      int(time.Now().Sub(time.Unix(330480000, 0)).Seconds()),
								},
								v3action.Instance{
									Index:       2,
									State:       "RUNNING",
									MemoryUsage: 3000000,
									DiskUsage:   3000000,
									MemoryQuota: 33554432,
									DiskQuota:   6000000,
									Uptime:      int(time.Now().Sub(time.Unix(1277164800, 0)).Seconds()),
								},
							},
						},
					},
				}
				fakeActor.GetApplicationSummaryByNameAndSpaceReturns(summary, v3action.Warnings{"warning-1", "warning-2"}, nil)
			})

			It("prints the application summary and outputs warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("(?m)Showing health and status for app some-app in org some-org / space some-space as steve\\.\\.\\.\n\n"))
				Expect(testUI.Out).To(Say("name:\\s+some-app"))
				Expect(testUI.Out).To(Say("requested state:\\s+started"))
				Expect(testUI.Out).To(Say("processes:\\s+web:3/3, console:0/0, worker:0/1"))
				Expect(testUI.Out).To(Say("memory usage:\\s+32M x 3, 64M x 1"))
				Expect(testUI.Out).To(Say("stack:\\s+cflinuxfs2"))
				Expect(testUI.Out).To(Say("(?m)buildpacks:\\s+some-detect-output, some-buildpack\n\n"))
				Expect(testUI.Out).To(Say("web"))
				Expect(testUI.Out).To(Say("\\s+state\\s+since\\s+cpu\\s+memory\\s+disk"))
				Expect(testUI.Out).To(Say("#0\\s+running\\s+1978-\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2} [AP]M\\s+0.0%\\s+976.6K of 32M\\s+976.6K of 1.9M"))
				Expect(testUI.Out).To(Say("#1\\s+running\\s+1980-\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2} [AP]M\\s+0.0%\\s+1.9M of 32M\\s+1.9M of 3.8M"))
				Expect(testUI.Out).To(Say("#2\\s+running\\s+2010-\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2} [AP]M\\s+0.0%\\s+2.9M of 32M\\s+2.9M of 5.7M"))

				Expect(testUI.Out).To(Say("console"))
				Expect(testUI.Out).To(Say("There are no running instances of this process."))

				Expect(testUI.Out).To(Say("worker"))
				Expect(testUI.Out).To(Say("There are no running instances of this process."))

				Expect(testUI.Err).To(Say("warning-1"))
				Expect(testUI.Err).To(Say("warning-2"))

				Expect(fakeActor.GetApplicationSummaryByNameAndSpaceCallCount()).To(Equal(1))
				appName, spaceGUID := fakeActor.GetApplicationSummaryByNameAndSpaceArgsForCall(0)
				Expect(appName).To(Equal("some-app"))
				Expect(spaceGUID).To(Equal("some-space-guid"))
			})
		})
	})

	Context("when the actor returns an error", func() {
		var expectedErr error

		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				Name: "some-space",
			})
			fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
			expectedErr = v3action.ApplicationNotFoundError{Name: app}
			fakeActor.GetApplicationSummaryByNameAndSpaceReturns(v3action.ApplicationSummary{}, v3action.Warnings{"warning-1", "warning-2"}, expectedErr)
		})

		It("returns the error and prints warnings", func() {
			Expect(executeErr).To(Equal(command.ApplicationNotFoundError{Name: app}))

			Expect(testUI.Out).To(Say("Showing health and status for app some-app in org some-org / space some-space as steve\\.\\.\\."))

			Expect(testUI.Err).To(Say("warning-1"))
			Expect(testUI.Err).To(Say("warning-2"))
		})
	})
})
