# Configuration Guide for gh-stars

This guide provides detailed instructions on how to configure gh-stars, including setting up the `.netrc` file with a GitHub personal access token.

## Prerequisites

Before configuring gh-stars, ensure you have:

1. Installed gh-stars on your system (refer to the README.md for installation instructions)
2. A GitHub account
3. Basic familiarity with command-line operations

## Setting up the .netrc file

gh-stars uses a `.netrc` file to store your GitHub authentication credentials. Follow these steps to set up your `.netrc` file:

1. Create or edit the `.netrc` file in your home directory:

   ```bash
   nano ~/.netrc
   ```

2. Add the following content to the file:

   ```
   machine api.github.com
       login YOUR_GITHUB_USERNAME
       password YOUR_GITHUB_PERSONAL_ACCESS_TOKEN
   ```

   Replace `YOUR_GITHUB_USERNAME` with your actual GitHub username and `YOUR_GITHUB_PERSONAL_ACCESS_TOKEN` with your personal access token.

3. Save the file and exit the editor.

4. Secure the file by setting appropriate permissions:

   ```bash
   chmod 600 ~/.netrc
   ```

## Obtaining a GitHub Personal Access Token

To use gh-stars, you need a GitHub personal access token. Here's how to create one:

1. Log in to your GitHub account.
2. Go to Settings > Developer settings > Personal access tokens.
3. Click "Generate new token".
4. Give your token a descriptive name.
5. Select the necessary scopes for gh-stars (typically, you'll need at least `repo` and `user` scopes).
6. Click "Generate token".
7. Copy the generated token immediately (you won't be able to see it again).

## Verifying the Configuration

To verify that gh-stars is correctly configured:

1. Run the following command:

   ```bash
   stars show
   ```

2. If configured correctly, you should see a list of your starred repositories without any authentication errors.

## Troubleshooting

If you encounter issues:

- Ensure your `.netrc` file has the correct permissions (600).
- Verify that your GitHub username and personal access token are correct in the `.netrc` file.
- Check that your personal access token has the necessary scopes.

For more detailed information on using gh-stars, refer to the README.md file in the project repository.