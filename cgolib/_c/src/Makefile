NO_PID_API ?= y
export NO_PID_API

# XXX: modify as desired
PREFIX ?= ../../../libs
export PREFIX

ifdef DEBUG
export DEBUG
endif

export DEBUG=y
export SHARED=n

.PHONY: all clean TAGS install uninstall style cppcheck

SUBDIRS = \
    intel-cmt-cat \
    $(NULL)

# include $(top_srcdir)/build-aux/Makefile.subs

install-lib:
	$(MAKE) -C intel-cmt-cat/lib install
lib:
	$(MAKE) -C intel-cmt-cat/lib

clean-lib:
	$(MAKE) -C intel-cmt-cat/lib clean

uninstall-lib:
	$(MAKE) -C intel-cmt-cat/lib uninstall

TAGS:
	find ./ -name "*.[ch]" -print | etags -
