const webpack = require("webpack");
const { merge } = require("webpack-merge");
const terser = require("terser-webpack-plugin");
const common = require("./webpack.common.js");

module.exports = merge(common, {
  mode: "production",
  devtool: "source-map",
  optimization: {
    minimize: true,
    minimizer: [new terser()],
  },
  plugins: [
    new webpack.DefinePlugin({
      "process.env.NODE_ENV": JSON.stringify("production")
    })
  ]
});
