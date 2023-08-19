PLUGIN_SRC_DIR = ./plugins/source
PLUGIN_DST_DIR = ./plugins/compiled

# List all directories inside ./plugins/source/
PLUGIN_DIRS = $(wildcard $(PLUGIN_SRC_DIR)/*)
# Convert the directory names into expected .so names in ./plugins/compiled
PLUGIN_OUT = $(patsubst $(PLUGIN_SRC_DIR)/%, $(PLUGIN_DST_DIR)/%.so, $(PLUGIN_DIRS))

rebuild: clean all

all:  $(PLUGIN_OUT)

# Rule to compile .go files in sub-directories into .so files
$(PLUGIN_DST_DIR)/%.so: $(PLUGIN_SRC_DIR)/%
	@mkdir -p $(PLUGIN_DST_DIR)
	go build -o $@ -buildmode=plugin $</*.go

clean:
	rm -f $(PLUGIN_DST_DIR)/*.so

