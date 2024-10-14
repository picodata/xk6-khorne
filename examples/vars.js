export default {
    hosts: __ENV.HOSTS ? __ENV.HOSTS.split(",") : ["127.0.0.1:8010"],
    generalOpts: {
      summaryTrendStats: ["min", "avg", "p(95)", "max", "count"],
      systemTags: ["status", "check", "error", "error_code"],
      thresholds: {
        checks: ["rate>0.99"],
      },
    },
  };
  