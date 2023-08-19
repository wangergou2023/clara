# AI Assistant with OpenAI & Plugin Integration

Extensible AI assistant that seamlessly interacts with various plugins to fulfill complex tasks, all through natural language conversations powered by OpenAI.

## Table of Contents
1. [Key Features](#key-features)
2. [Installation & Setup](#installation--setup)
3. [Usage](#usage)
4. [Known Issues](#known-issues)
5. [Future Enhancements & Missing Features](#future-enhancements--missing-features)
6. [Contribution](#contribution)
7. [License](#license)

## Key Features

- **Dynamic Plugin Integration**: Easily extend the capabilities of the AI assistant by adding new plugins.
  
- **Function Chaining**: Execute multiple plugins in sequence for more complex commands.
  
- **Conversational Interface**: Color-coded roles and an interactive design ensure clear communication.

- **Persistant long term memory**: The assistant remembers information from previous conversations to provide a more personalized experience. via a Vector database (Milvus)

![Memory example](Clara-memorygif.gif)

## Installation & Setup

1. Clone the repository:
```bash
git clone https://github.com/jjkirkpatrick/clara
```

2. Navigate to the project directory and install the required packages:
```bash
cd clara
make buildAssistant
```

3. Set up the OpenAI API key and other configurations as required.

4. Run the assistant:
```bash
./clara
```

If you don't wish to build the assistant binary you can run it directly with:
```bash
go run main.go
```
However, you will still need to compile the plugins with `make rebuild` before running the assistant.

## Usage

Start a conversation with the assistant by typing `./clara` in the terminal. You can then enter commands in natural language to interact with the assistant.

You can ask the assistant what functions is has available by using natural language commands such as:
- "What are your functions?"
- "What can I ask you to do?"


## Known Issues

- _[List any known bugs or issues here.]_

## Future Enhancements & Missing Features

- support for the assistant to create and load it's own plugins at runtime if it doesn't have a plugin for a given function (e.g. if the user asks the assistant to do something it doesn't know how to do, it can create a plugin for that function and then execute it)

## Contribution

Contributions are always welcome! See [CONTRIBUTING.md](./CONTRIBUTING.md) for ways to get started. Please adhere to this project's [code of conduct](./CODE_OF_CONDUCT.md).

## License

This project is licensed under the MIT License
