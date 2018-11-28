const path = require('path');
const CleanWebpackPlugin = require('clean-webpack-plugin');

module.exports = {
  entry: {
    login: './public/src/js/login.js',
    register: './public/src/js/register.js',
    'admin-search': './public/src/js/admin-search.js',
    'admin-users': './public/src/js/admin-users.js',
    'admin-user': './public/src/js/admin-user.js',
    'manage-admin': './public/src/js/manage-admin.js',
    'manage-device': './public/src/js/manage-device.js',
    manage: './public/src/js/manage.js',
    captcha: './public/src/js/captcha.js'
  },

  plugins: [
    new CleanWebpackPlugin(['public/dist/js'])
  ],

  output: {
    filename: '[name].min.js',
    path: path.resolve(__dirname, 'public/dist/js')
  },

  resolve: {
    modules: [
      path.resolve(__dirname, "public/src/modules"),
      "node_modules"
    ]
  },

  module: {
    rules: [
      {
        test: /\.js$/,
        exclude: /node_modules/,
        use: {
          loader: 'babel-loader'
        }
      }
    ]
  }
};
