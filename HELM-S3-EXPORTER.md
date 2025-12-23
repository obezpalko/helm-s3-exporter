# helm s3 exporter

## key features

- kubernetes application
- configurable s3 bucket
- configurable scan time (with timeout)
- (optional) create/update index.html with list of charts and versions from index.yaml (use icon parameter from chart if available)
  
### Authentication/Authorization

- by default use service account with role (configurable)
- if not - use secrets (existing/external/generated from values (warn that it is insecure))

## Proposed logic

- application connects to s3 bucket and get index.yaml with helm charts
- analyse obtained data and generate summary
  - total number of charts with d
  - number of versions of each chart
  - ages of each chart (oldest object date, newest object date, median date)
- expose this summary as prometheus metrics
- there is possibility to use configurable service and servicemonitor to use discovery feature of prometheus

