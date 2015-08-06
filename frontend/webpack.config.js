/* global require, __dirname, module */
var path = require('path');
var webpack = require('webpack');

module.exports = {
  entry: './src/index',
  output: {
    path: __dirname + '/static/',
    filename: 'sporedock.js',
    publicPath: '/scripts/'
  },
  plugins: [
    new webpack.NoErrorsPlugin(),
    new webpack.optimize.UglifyJsPlugin()
  ],
  resolve: {
    extensions: ['', '.js'],
    root: path.resolve(__dirname, 'src', 'js')
  },
  module: {
    loaders: [
      { test: /\.js$/, loaders: ['babel?stage=0'], exclude: /node_modules/ },
      { test: /\.scss$/, loaders: ['style', 'css', 'sass'], exclude: /node_modules/ },
      { test: /\.css$/, loaders: ['style', 'css'], exclude: /node_modules/ }
    ]
  }
};
