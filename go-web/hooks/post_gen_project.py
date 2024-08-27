import os

abs_path = os.getcwd() # this will respect to the output directory
project_name = '{{ cookiecutter.project_name }}'

for root, dirs, files in os.walk(abs_path):
    for filename in files:
        try:
            with open(os.path.join(root, filename)) as f:
                content = f.read()
                content = content.replace('exampleproj', project_name)
                with open(os.path.join(root, filename), 'w') as f:
                    f.write(content)

        except Exception as e:
            print(f'error reading file: {filename}, {e}')
