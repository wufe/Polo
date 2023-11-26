const path = require('path');
const ReactRefreshWebpackPlugin = require('@pmmmwh/react-refresh-webpack-plugin');
const webpack = require('webpack');
const TsconfigPathsPlugin = require('tsconfig-paths-webpack-plugin');
const ForkTsCheckerWebpackPlugin = require('fork-ts-checker-webpack-plugin');
const { CleanWebpackPlugin } = require('clean-webpack-plugin');
const CopyPlugin = require('copy-webpack-plugin');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const MiniCssExtractPlugin = require("mini-css-extract-plugin");
const CssMinimizerPlugin = require('css-minimizer-webpack-plugin');

const isProduction = process.env.NODE_ENV === 'production';
const isDevelopment = !isProduction;
const mode = isProduction ? 'production' : 'development';
const cwd = process.cwd();

module.exports = {
	mode,
	context: __dirname,
	name: 'session-helper',
	entry: {
		'session-helper': path.resolve(__dirname, 'src', 'index.tsx'),
	},
	output: {
		publicPath: '/_polo_/public/',
		filename: isDevelopment ? '[name].js' : '[name].[fullhash].js',
		chunkFilename: isDevelopment ? '[id]-chunk.js' : '[id]-chunk.[fullhash].js',
		path: path.resolve(cwd, 'pkg/services/static')
	},
	devtool: 'source-map',
	resolve: {
		extensions: ['.ts', '.tsx', '.js', '.scss', '.sass', '.hbs'],
		plugins: [new TsconfigPathsPlugin({ configFile: path.resolve(__dirname, 'tsconfig.json') })],
		alias: {
			'react': 'preact/compat',
			'react-dom': 'preact/compat',
		}
	},
	module: {
		rules: [
			{
				test: /\.tsx?$/,
				exclude: /node_modules/,
				use: [
					{
						loader: 'ts-loader',
						options: {
							transpileOnly: true,
							compilerOptions: {
								"jsx": "react-jsx",
								"jsxImportSource": "preact"
							}
						}
					},
				]

			},
			{
				test: /\.s[ac]ss$/,
				use: [
					MiniCssExtractPlugin.loader,
					'css-loader',
					{
						loader: 'sass-loader',
						options: {
							sassOptions: {
								includePaths: [
									"node_modules"
								]
							}
						}
					}
				]
			},
			{ test: /\.hbs$/, loader: 'handlebars-loader' },
			{ test: /\.(png|svg)$/, loader: 'file-loader' },

		]
	},
	target: 'web',
	optimization: {
		minimizer: [
			"...",
			isProduction && new CssMinimizerPlugin(),
		].filter(Boolean)
	},
	plugins: [
		new ForkTsCheckerWebpackPlugin({
			typescript: {
				configFile: path.resolve(__dirname, 'tsconfig.json')
			}
		}),
		new MiniCssExtractPlugin({
			// Options similar to the same options in webpackOptions.output
			// both options are optional
			filename: isDevelopment ? '[name].css' : '[name].[fullhash].css',
			chunkFilename: isDevelopment ? '[id]-chunk.css' : '[id].[fullhash].css',
		}),
		new HtmlWebpackPlugin({
			filename: './session-helper.html',
			template: path.resolve(__dirname,'src', 'index.hbs'),
			inject: false,
			chunks: ['session-helper']
		}),

	].filter(Boolean),
}