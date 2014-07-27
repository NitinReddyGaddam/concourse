package builder_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/tedsuo/rata"

	TurbineBuilds "github.com/concourse/turbine/api/builds"
	TurbineRoutes "github.com/concourse/turbine/routes"

	WinstonRoutes "github.com/concourse/atc/api/routes"
	. "github.com/concourse/atc/builder"
	"github.com/concourse/atc/builder/fakes"
	"github.com/concourse/atc/builds"
	"github.com/concourse/atc/config"
)

var _ = Describe("Builder", func() {
	var db *fakes.FakeBuilderDB

	var turbineServer *ghttp.Server

	var build builds.Build

	var builder Builder

	var job config.Job
	var resources config.Resources

	var expectedTurbineBuild TurbineBuilds.Build

	BeforeEach(func() {
		db = new(fakes.FakeBuilderDB)

		turbineServer = ghttp.NewServer()

		job = config.Job{
			Name: "some-job",

			BuildConfig: TurbineBuilds.Config{
				Image: "some-image",
				Params: map[string]string{
					"FOO": "1",
					"BAR": "2",
				},
				Run: TurbineBuilds.RunConfig{
					Path: "some-script",
					Args: []string{"arg1", "arg2"},
				},
			},

			Privileged: true,

			BuildConfigPath: "some-resource/build.yml",

			Inputs: []config.Input{
				{
					Resource: "some-resource",
					Params:   config.Params{"some": "params"},
				},
			},
		}

		resources = config.Resources{
			{
				Name:   "some-resource",
				Type:   "git",
				Source: config.Source{"uri": "git://some-resource"},
			},
			{
				Name:   "some-dependant-resource",
				Type:   "git",
				Source: config.Source{"uri": "git://some-dependant-resource"},
			},
			{
				Name:   "some-output-resource",
				Type:   "git",
				Source: config.Source{"uri": "git://some-output-resource"},
			},
		}

		expectedTurbineBuild = TurbineBuilds.Build{
			Config: TurbineBuilds.Config{
				Image: "some-image",

				Params: map[string]string{
					"FOO": "1",
					"BAR": "2",
				},

				Run: TurbineBuilds.RunConfig{
					Path: "some-script",
					Args: []string{"arg1", "arg2"},
				},
			},

			Inputs: []TurbineBuilds.Input{
				{
					Name:       "some-resource",
					Type:       "git",
					Source:     TurbineBuilds.Source{"uri": "git://some-resource"},
					Params:     TurbineBuilds.Params{"some": "params"},
					ConfigPath: "build.yml",
				},
			},

			Outputs: []TurbineBuilds.Output{},

			Privileged: true,

			Callback: "http://atc-server/builds/some-job/128",
			LogsURL:  "ws://atc-server/builds/some-job/128/log/input",
		}

		builder = NewBuilder(
			db,
			resources,
			rata.NewRequestGenerator(turbineServer.URL(), TurbineRoutes.Routes),
			rata.NewRequestGenerator("http://atc-server", WinstonRoutes.Routes),
		)

		build = builds.Build{
			ID: 128,
		}
	})

	successfulBuildStart := func(build TurbineBuilds.Build) http.HandlerFunc {
		createdBuild := build
		createdBuild.Guid = "some-turbine-guid"
		createdBuild.AbortURL = turbineServer.URL() + "/abort/the/build"

		return ghttp.CombineHandlers(
			ghttp.VerifyJSONRepresenting(build),
			ghttp.RespondWithJSONEncoded(201, createdBuild),
		)
	}

	Context("when the build can be scheduled", func() {
		BeforeEach(func() {
			db.ScheduleBuildReturns(true, nil)
			db.StartBuildReturns(true, nil)
		})

		It("starts the build and saves its abort url", func() {
			turbineServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/builds"),
					successfulBuildStart(expectedTurbineBuild),
				),
			)

			err := builder.Build(build, job, nil)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(db.StartBuildCallCount()).Should(Equal(1))

			job, id, abortURL := db.StartBuildArgsForCall(0)
			Ω(job).Should(Equal("some-job"))
			Ω(id).Should(Equal(128))
			Ω(abortURL).Should(ContainSubstring("/abort/the/build"))
		})

		Context("when the build fails to transition to started", func() {
			BeforeEach(func() {
				db.StartBuildReturns(false, nil)
			})

			It("aborts the build on the turbine", func() {
				turbineServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/builds"),
						successfulBuildStart(expectedTurbineBuild),
					),
					ghttp.VerifyRequest("POST", "/abort/the/build"),
				)

				err := builder.Build(build, job, nil)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(turbineServer.ReceivedRequests()).Should(HaveLen(2))
			})
		})

		Context("when the build has outputs", func() {
			BeforeEach(func() {
				job.Outputs = []config.Output{
					{
						Resource: "some-resource",
						Params:   config.Params{"foo": "bar"},
					},
				}

				expectedTurbineBuild.Outputs = []TurbineBuilds.Output{
					{
						Name:       "some-resource",
						Type:       "git",
						Params:     TurbineBuilds.Params{"foo": "bar"},
						SourcePath: "some-resource",
						Source:     TurbineBuilds.Source{"uri": "git://some-resource"},
					},
				}
			})

			It("sends them along to the turbine", func() {
				turbineServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/builds"),
						successfulBuildStart(expectedTurbineBuild),
					),
				)

				err := builder.Build(build, job, nil)
				Ω(err).ShouldNot(HaveOccurred())
			})
		})

		Context("when versioned resources are specified", func() {
			BeforeEach(func() {
				expectedTurbineBuild.Inputs = []TurbineBuilds.Input{
					{
						Name:       "some-resource",
						Type:       "git-ng",
						Source:     TurbineBuilds.Source{"uri": "git://some-provided-uri"},
						Params:     TurbineBuilds.Params{"some": "params"},
						Version:    TurbineBuilds.Version{"version": "1"},
						ConfigPath: "build.yml",
					},
				}
			})

			It("uses them for the build's inputs", func() {
				turbineServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/builds"),
						successfulBuildStart(expectedTurbineBuild),
					),
				)

				err := builder.Build(build, job, builds.VersionedResources{
					{
						Name:    "some-resource",
						Type:    "git-ng",
						Version: builds.Version{"version": "1"},
						Source:  config.Source{"uri": "git://some-provided-uri"},
					},
				})
				Ω(err).ShouldNot(HaveOccurred())
			})
		})

		Context("when the job's input is not found", func() {
			BeforeEach(func() {
				job.Inputs = append(job.Inputs, config.Input{
					Resource: "some-bogus-resource",
				})
			})

			It("returns an error", func() {
				err := builder.Build(build, job, nil)
				Ω(err).Should(HaveOccurred())
			})
		})

		Context("when the job's output is not found", func() {
			BeforeEach(func() {
				job.Outputs = append(job.Outputs, config.Output{
					Resource: "some-bogus-resource",
				})
			})

			It("returns an error", func() {
				err := builder.Build(build, job, nil)
				Ω(err).Should(HaveOccurred())
			})
		})

		Context("when the turbine server is unreachable", func() {
			BeforeEach(func() {
				turbineServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/builds"),
						func(w http.ResponseWriter, r *http.Request) {
							turbineServer.HTTPTestServer.CloseClientConnections()
						},
					),
				)
			})

			It("returns an error", func() {
				err := builder.Build(build, job, nil)
				Ω(err).Should(HaveOccurred())

				Ω(turbineServer.ReceivedRequests()).Should(HaveLen(1))
			})
		})

		Context("when the turbine server returns non-201", func() {
			BeforeEach(func() {
				turbineServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/builds"),
						ghttp.RespondWith(400, ""),
					),
				)
			})

			It("returns an error", func() {
				err := builder.Build(build, job, nil)
				Ω(err).Should(HaveOccurred())
			})
		})
	})

	Context("when the build cannot be scheduled", func() {
		BeforeEach(func() {
			db.ScheduleBuildReturns(false, nil)
		})

		Context("and the current build is scheduled", func() {
			It("leaves the build pending", func() {
				err := builder.Build(build, job, nil)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(db.StartBuildCallCount()).Should(Equal(0))
			})

			It("does not trigger a build", func() {
				err := builder.Build(build, job, nil)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(turbineServer.ReceivedRequests()).Should(BeEmpty())
			})
		})
	})
})
