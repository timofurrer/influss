module.exports = {
  ignoreFiles: [
    'package.json',
    'package-lock.json',
    'web-ext-config.js',
    'README.md',
    '.git',
    '.github',
    'node_modules'
  ],
  build: {
    overwriteDest: true
  },
  sign: {
    channel: 'listed'
  }
};
