[![Build and Test](https://github.com/danielchalef/mrfparse/actions/workflows/main.yml/badge.svg)](https://github.com/danielchalef/mrfparse/actions/workflows/main.yml)

**This repo is no longer maintained**

# A Go parser for _Transparency in Coverage_ MRF files.
`mrfparse` is a memory and CPU efficient parser for _Transparency in Coverage_ Machine Readable Format (MRF) files. The parser is designed to be easily containerized and scaled on modern cloud container platforms (and potentially cloud function infrastructure).

`mrfparse` is fast: Parsing out pricing and providers for the CMS' _500 shoppable services_ from an 80GB Anthem _in-network-rates_ fileset in NDJSON format to parquet takes <5 minutes on a 12-core workstation with container memory limited to 6GB. Doing the same from the gzip compressed source file takes an additional ~5 minutes.

Features:

- Outputs to a parquet dataset, allowing easy ingestion into data warehouses and data lakes.
- Supports reading from HTTP, and S3 / GS cloud storage, and writing to S3 / GS cloud storage buckets.
- Filter for a subset of CPT/HCPCS service codes (provided as a simple CSV file).
- Filters for only providers for whom pricing data is present in the MRF file, dropping extranous provider data.
- Supports reading Gzip compressed MRF files.
- The output schema is designed to support ingestion into graph databases.

## Background
As of July 1, 2022, _The Centers for Medicare and Medicaid Services (CMS)_ mandated that most group health plans and issuers of group or individual health insurance (payers) [must post pricing information for covered items and services](https://www.cms.gov/healthplan-price-transparency/public-data). The data is available in a machine readable format (MRF) that is described in the [Transparency in Coverage](https://github.com/CMSgov/price-transparency-guide) Github repo.

Working with MRF files is challenging:
- Each payer's MRF dataset is tens to hundreds of terabytes of data and is updated monthly. No monthly deltas are available and individual JSON documents can be over 1TB in size.
- Some payers have included provider data for providers for whom the MRF file does not have pricing data. That is, there are provider reference records where in_network rates are not present.
- Some payers have provided pricing data for services that providers do not offer.

## Usage
The following examples illustrate using the binary from a command line. 


Parse a gzipped MRF file hosted on a payer's website and output the parquet dataset to an S3 bucket
```bash
mrfparse pipeline -i https://mrf.healthsparq.com/aetnacvs/inNetworkRates/2022-12-05_Innovation-Health-Plan-Inc.json.gz \
                  -o s3://mrfdata/staging/2022-12-05/aetnacvs/ \
                  -p 99
```


Parse a gzipped MRF file hosted in a Google Cloud Storage bucket and output the parquet dataset to the local filesystem.
```bash
mrfparse pipeline -i gs://mrfdata/staging/2022-12-05_Innovation-Health-Plan-Inc.json.gz \
                  -o mrfdata/staging/2022-12-05/aetnacvs/ \
                  -p 99
```

`mrfparse` operates in several stages each of which can be executed independently. See `mrfparse --help` for more options.

### Production Use
It is strongly recommended that you use the containerized parser and run it on a cloud container platform, allowing many files to be parsed concurrenlty. The "all-in-one" `pipeline` is not recommended for production use. For more resilient data pipelines, it is recommended that you use something like Airflow to run each of the download, `split` and `parse` steps sequentially in a recoverable way.

Additionally, see the note below regarding not using `mrfparse` on ARM64 processors in production.

## Requirements
`mrfparse` makes extensive use of [`simdjson-go`](https://github.com/minio/simdjson-go) to parse MRF JSON documents. A CPU with both AVX2 and CLMUL instruction support is required (most modern Intel or AMD processors). Unfortunately, `simdjson-go` does not (yet) support ARM64 NEON.

Other requirements:
- 6GB of RAM (though I'd like to reduce this)
- Adequate temporal storage for intermediate data files.

### Note on ARM Compatibility
To enable local testing with non-amd64 cpu's, such as Apple's new M# series of machines, this utility makes use of the
[fakesimdjson](https://github.com/kiwicom/fakesimdjson) package.  When using this simdjson simulacrum parsing speed and
efficiency will be drastically reduced. It is therefore _not_ recommended to use this on ARM-based machines in a 
production environment.

## Build and Installation
Using `go install`:
```bash
go install github.com/danielchalef/mrfparse@latest
```

Use the `Makefile` to build the binary or container. 

Build the binary
```bash
make 
```

Build the container
```bash
make docker-build
```

Edit the `Makefile` to change the container registry and tag and then release to your registry:
```bash
make docker-release
```

See `make help` for more options.

## Configuration and Tuning

### Configuration via `config.yml` and environment variables

A number of runtime options can be set via a `config.yml` file. The default location is `./config.yml`. The location can be changed via the `--config` flag. These options may also be set via environment variables prefixed with `MRF_`.
```yaml
log:
  level: info
services:
  file: services.csv
writer:
  max_rows_per_file: 100_000_000
  filename_template: "_%04d.zstd.parquet"
  max_rows_per_group: 1_000_000
tmp:
  path: /tmp
pipeline:
  download_timeout: 20          # minutes
```

### The `services` file
`mrfparse` is designed to parse out only a selected list of services identified by CPT/HCPCS codes. This list of codes needs to be provided to `mrfparse` in the form of a simple `csv` file which may be on a local filesystem or hosted on S3/GS. 

Use either the `config.yaml` file or the `--services` flag to specify the location of the `services` file. The default location is `./services.csv`. A sample services file containing the CMS' _500 Shoppable Services_ may be found in the `data` folder in this repo.

### Tuning
UPDATE: `jsplit` now makes use of pooled buffers and is much faster than it was when this was written. YMMV on the following.

Splitting an MRF JSON document into NDJSON using `jsplit` takes time. `jsplit` makes heavy usage of the GC and can be sped up by setting a `GOGC` value far higher than the default of 200, at the expense of a non-linear increase in memory usage.

## Parquet Schema

See the models in [`models/mrf.go`](pkg/mrfparse/models/mrf.go) for the parquet schema.

## How the core parser works
An MRF file is split into a set of JSON documents using a fork of [`jsplit`](https://github.com/dolthub/jsplit) that has been modified to support reading and writing to cloud storage and use as a Go module. `jsplit` generates a root document and set of `provider-reference` and `in-network-rates` files. These files are in NDJSON format, allowing them to be consumed memory efficently. They are parsed line by line using [`simdjson-go`](https://github.com/minio/simdjson-go) and output to a parquet dataset.

`in-network-rates` files are parsed first, allowing us to filter against our `services` list and build up a list of providers for whom we have pricing data. This provider list is then used to filter the `provider-reference` files. 

## Status
- Currently, only [in-network-rates](https://github.com/CMSgov/price-transparency-guide/tree/master/schemas/in-network-rates) files are supported. 
- Providers are indentified by either their NPI number or EIN. No effort has been made to enrich the data with additional provider information (e.g. provider name, address, etc.).
- The parser does not attempt to validate that a provider actually provides a specific service that the MRF file offers pricing for.
- `mrfparse` is not a validating parser but does attempt to detect and report some errors in the MRF file. Note that payers _do_ deviate from the CMS' schema!
- The parser has been extensively tested with Anthem and Aetna datasets. YMMV with other payers.

Contributions and feedback are welcome. This was my first large-ish Go project. Please do let me know if you have any suggestions for improvement.

## Acknowledgments
- [simdjson-go](https://github.com/minio/simdjson-go)
- [jsplit](https://github.com/dolthub/jsplit)

## License
This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.

   Copyright 2023 Daniel Chalef

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
