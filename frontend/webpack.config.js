/* global require, __dirname, module */
var path = require('path');
var webpack = require('webpack');

module.exports = {
  entry: './src/index',
  output: {
    path: __dirname + '/scripts/',
    filename: 'sporedock.js',
    publicPath: '/scripts/'
  },
  plugins: [
    new webpack.NoErrorsPlugin()
  ],
  resolve: {
    extensions: ['', '.js'],
    root: path.resolve(__dirname, 'src', 'js')
  },
  module: {
    loaders: [
      { test: /\.js$/, loaders: ['babel'], exclude: /node_modules/ },
      { test: /\.scss?$/, loaders: ['style', 'css', 'sass'], exclude: /node_modules/ }
    ]
  }
};
