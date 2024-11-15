const path = require("path");
const { CleanWebpackPlugin } = require("clean-webpack-plugin");

module.exports = {
  entry: {
    login: "./public/src/ts/login.ts",
    register: "./public/src/ts/register.ts",
    "flash-messages": "./public/src/ts/flash-messages.ts",
    "admin-users": "./public/src/ts/admin-users.ts",
    "admin-user": "./public/src/ts/admin-user.ts",
    "manage-admin": "./public/src/ts/manage-admin.ts",
    "manage-device": "./public/src/ts/manage-device.ts",
    manage: "./public/src/ts/manage.ts",
    captcha: "./public/src/ts/captcha.ts"
  },

  plugins: [new CleanWebpackPlugin()],

  output: {
    filename: "[name].min.js",
    path: path.resolve(__dirname, "public/dist/js")
  },

  resolve: {
    extensions: [".ts", ".js"],
    modules: [path.resolve(__dirname, "public/src/modules"), "node_modules"],
    alias: {
      "@": path.resolve(__dirname, "public/src/modules/")
    }
  },

  module: {
    rules: [
      {
        test: /\.ts$/,
        exclude: /node_modules/,
        use: {
          loader: "ts-loader"
        }
      }
    ]
  }
};
