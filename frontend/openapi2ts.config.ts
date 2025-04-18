const path = require('path');

export default {
  schemaPath: path.join('http://11.161.204.49:30871', '/swagger/doc.json'),
  serversPath: './src/service',
  requestLibPath: '@/util/request',
  projectName: 'obshell',
};
