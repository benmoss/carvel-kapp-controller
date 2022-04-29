package hack

_oses: ["linux"]
_arches: ["arm64", "amd64"]

#Dependency: {
	name:    string
	version: string
	artifacts: {for os in _oses {
		for arch in _arches {
			"\(os)": "\(arch)": #Artifact
		}
	}}
}

#Artifact: {
	url:             string
	checksum:        string
	tarballSubpath?: string
}

#CarvelDependency: #Dependency & {
	name:    string
	version: string
	artifacts: [os=string]: [arch=string]: url: "https://github.com/vmware-tanzu/carvel-\(name)/releases/download/\(version)/\(name)-\(os)-\(arch)"
}

ytt: #CarvelDependency & {
	name:    "ytt"
	version: "v0.40.1"
	artifacts: {
		linux: arm64: checksum: "c5d2f033b375ee87414b51d16c324d7a441de2f256865b5c774c4d5aea30ef60"
		linux: amd64: checksum: "11222665c627b8f0a1443534a3dde3c9b3aac08b322d28e91f0e011e3aeb7df5"
	}
}

kbld: #CarvelDependency & {
	name:    "kbld"
	version: "v0.32.0"
	artifacts: {
		linux: amd64: checksum: "de546ac46599e981c20ad74cd2deedf2b0f52458885d00b46b759eddb917351a"
		linux: arm64: checksum: "4e9995a086f73ed1f7b2c0d31e139ccdd82c243e5f9af133bfd0dda7d8e5a063"
	}
}
kapp: #CarvelDependency & {
	name:    "kapp"
	version: "v0.46.0"
	artifacts: {
		linux: amd64: checksum: "130f648cd921761b61bb03d7a0f535d1eea26e0b5fc60e2839af73f4ea98e22f"
		linux: arm64: checksum: "f40cf819f170ce50632bcb683098a05711d0a2a7110b72eb2754a7fc651eb2f3"
	}
}
imgpkg: #CarvelDependency & {
	name:    "imgpkg"
	version: "v0.28.0"
	artifacts: {
		linux: amd64: checksum: "8d22423dd6d13efc0e580443d8f88d2183c52c6f851ba51e3e54f25bf140be58"
		linux: arm64: checksum: "53ea4a9eec4bb1ff6adb701018da9978aad45bdee161d68e08bc69d84459f2c8"
	}
}

vendir: #CarvelDependency & {
	name:    "vendir"
	version: "v0.26.0"
	artifacts: {
		linux: amd64: checksum: "98057bf90e09972f156d1c4fbde350e94133bbaf2e25818b007759f5e9c8b197"
		linux: arm64: checksum: "d9f5b6e1438d87167863bf744d30c8a40f6bbea018f56ea51a9baf57fcf3609a"
	}
}

helm: #Dependency & {
	name:    "helm"
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
	version: "v3.7.2"
	artifacts: {
		[os=string]: [arch=string]: url: "https://github.com/mozilla/sops/releases/download/\(version)/sops-\(version).\(os).\(arch)"
		linux: arm64: checksum:          "86a6c48ec64255bd317d7cd52c601dc62e81be68ca07cdeb21a1e0809763647f"
		linux: amd64: checksum:          "0f54a5fc68f82d3dcb0d3310253f2259fef1902d48cfa0a8721b82803c575024"
	}
}

age: #Dependency & {
	name:    "age"
	version: "v1.0.0"
	artifacts: {
		[os=string]: [arch=string]: {
			url:            "https://github.com/FiloSottile/age/releases/download/\(version)/age-\(version)-\(os)-\(arch).tar.gz"
			tarballSubpath: "age/age"
		}
		linux: arm64: checksum: "6c82aa1d406e5a401ec3bb344cd406626478be74d5ae628f192d907cd78af981"
		linux: amd64: checksum: "6414f71ce947fbbea1314f6e9786c5d48436ebc76c3fd6167bf018e432b3b669"
	}
}

cue: #Dependency & {
	name:    "cue"
	version: "v0.4.2"
	artifacts: {
		[os=string]: [arch=string]: {
			url:            "https://github.com/cue-lang/cue/releases/download/\(version)/cue_\(version)_\(os)_\(arch).tar.gz"
			tarballSubpath: "cue"
		}
		linux: arm64: checksum: "6515c1f1b6fc09d083be533019416b28abd91e5cdd8ef53cd0719a4b4b0cd1c7"
		linux: amd64: checksum: "d43cf77e54f42619d270b8e4c1836aec87304daf243449c503251e6943f7466a"
	}
}
