# COPR build integration
# This Makefile is called by COPR to build the SRPM
#
# COPR will call: make -f .copr/Makefile srpm outdir="<resultdir>" spec="<spec_file>"
#
# For more info: https://docs.pagure.org/copr.copr/user_documentation.html#make-srpm

srpm:
	dnf install -y golang rpm-build
	$(MAKE) srpm SRPM_DIR=$(outdir)
