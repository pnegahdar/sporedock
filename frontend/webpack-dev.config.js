var config = require('./webpack.config');

/* global require, __dirname, module */
var webpack = require('webpack');

config.plugins = [
    new webpack.NoErrorsPlugin()
];

module.exports = config;
