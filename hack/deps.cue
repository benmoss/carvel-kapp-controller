package hack

_oses: ["linux", "darwin"]
_arches: ["arm64", "amd64"]

#Dependency: {
	name:    string
	version: string
	repo:    string
	dev:     *false | bool
	artifacts: {for os in _oses {
		for arch in _arches {
			"\(os)": "\(arch)": #Artifact
		}
	}}
}

#Artifact: {
	url:             string
	checksum?:       string
	tarballSubpath?: string
}

#CarvelDependency: #Dependency & {
	name:    string
	version: string
	dev:     true
	repo:    "vmware-tanzu/carvel-\(name)"
	artifacts: [os=string]: [arch=string]: url: "https://github.com/\(repo)/releases/download/\(version)/\(name)-\(os)-\(arch)"
}

ytt: #CarvelDependency & {
	name:    "ytt"
	version: "v0.40.1"
	artifacts: {
		linux: arm64: checksum:  "c5d2f033b375ee87414b51d16c324d7a441de2f256865b5c774c4d5aea30ef60"
		linux: amd64: checksum:  "11222665c627b8f0a1443534a3dde3c9b3aac08b322d28e91f0e011e3aeb7df5"
		darwin: amd64: checksum: "d46dba5e729e2fe36c369e96eaa2eb5354fb1bf7cf9184f9bfa829b8e5558b94"
	}
}

kbld: #CarvelDependency & {
	name:    "kbld"
	version: "v0.33.0"
	artifacts: {
		linux: amd64: checksum:  "38a5dad7ed478d209c8206d95546989b2730c7fed914c78d85eed68a2233688e"
		linux: arm64: checksum:  "4e9995a086f73ed1f7b2c0d31e139ccdd82c243e5f9af133bfd0dda7d8e5a063"
		darwin: amd64: checksum: "181ac8be5652b54344617d90aa8e83fbb41756d1b4b99168fec85d8813b3c1b2"
	}
}
kapp: #CarvelDependency & {
	name:    "kapp"
	version: "v0.46.0"
	artifacts: {
		linux: amd64: checksum:  "130f648cd921761b61bb03d7a0f535d1eea26e0b5fc60e2839af73f4ea98e22f"
		linux: arm64: checksum:  "f40cf819f170ce50632bcb683098a05711d0a2a7110b72eb2754a7fc651eb2f3"
		darwin: amd64: checksum: "7a3e5235689a9cc6d0e85ba66db3f1e57ab65323d3111e0867771111d2b0c1a3"
	}
}
imgpkg: #CarvelDependency & {
	name:    "imgpkg"
	version: "v0.28.0"
	artifacts: {
		linux: amd64: checksum:  "8d22423dd6d13efc0e580443d8f88d2183c52c6f851ba51e3e54f25bf140be58"
		linux: arm64: checksum:  "53ea4a9eec4bb1ff6adb701018da9978aad45bdee161d68e08bc69d84459f2c8"
		darwin: amd64: checksum: "e43142fdb197a62844acb29cb619d513346aac3c23732a4d180c0ad974d9562e"
	}
}

vendir: #CarvelDependency & {
	name:    "vendir"
	version: "v0.27.0"
	artifacts: {
		linux: amd64: checksum:  "1aa12d070f2e91fcb0f4d138704c5061075b0821e6f943f5a39676d7a4709142"
		linux: arm64: checksum:  "015977ae54d85bf2366d7affb0d582fecf79737f0eb80fa8a66de9f66e877b61"
		darwin: amd64: checksum: "c26547097d67f21e129a25557d9d36c7c0e109afe130adff63d3c83ce9459ecc"
	}
}

helm: #Dependency & {
	name:    "helm"
	repo:    "helm/helm"
	version: "v3.7.1"
	artifacts: {
		[os=string]: [arch=string]: {
			url:            "https://get.helm.sh/helm-\(version)-\(os)-\(arch).tar.gz"
			tarballSubpath: "\(os)-\(arch)/helm"
		}
		linux: arm64: checksum: "57875be56f981d11957205986a57c07432e54d0b282624d68b1aeac16be70704"
		linux: amd64: checksum: "6cd6cad4b97e10c33c978ff3ac97bb42b68f79766f1d2284cfd62ec04cd177f4"
	}
}

sops: #Dependency & {
	name:    "sops"
	repo:    "mozilla/sops"
	version: "v3.7.2"
	artifacts: {
		[os=string]: [arch=string]: url: "https://github.com/\(repo)/releases/download/\(version)/sops-\(version).\(os).\(arch)"
		linux: arm64: checksum:          "86a6c48ec64255bd317d7cd52c601dc62e81be68ca07cdeb21a1e0809763647f"
		linux: amd64: checksum:          "0f54a5fc68f82d3dcb0d3310253f2259fef1902d48cfa0a8721b82803c575024"
	}
}

age: #Dependency & {
	name:    "age"
	repo:    "FiloSottile/age"
	version: "v1.0.0"
	artifacts: {
		[os=string]: [arch=string]: {
			url:            "https://github.com/\(repo)/releases/download/\(version)/age-\(version)-\(os)-\(arch).tar.gz"
			tarballSubpath: "age/age"
		}
		linux: arm64: checksum: "6c82aa1d406e5a401ec3bb344cd406626478be74d5ae628f192d907cd78af981"
		linux: amd64: checksum: "6414f71ce947fbbea1314f6e9786c5d48436ebc76c3fd6167bf018e432b3b669"
	}
}

cue: #Dependency & {
	name:    "cue"
	version: "v0.4.2"
	repo:    "cue-lang/cue"
	artifacts: {
		[os=string]: [arch=string]: {
			url:            "https://github.com/\(repo)/releases/download/\(version)/cue_\(version)_\(os)_\(arch).tar.gz"
			tarballSubpath: "cue"
		}
		linux: arm64: checksum: "6515c1f1b6fc09d083be533019416b28abd91e5cdd8ef53cd0719a4b4b0cd1c7"
		linux: amd64: checksum: "d43cf77e54f42619d270b8e4c1836aec87304daf243449c503251e6943f7466a"
	}
}
