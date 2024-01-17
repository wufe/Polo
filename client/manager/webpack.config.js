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
	name: 'manager',
	entry: {
		'manager': './src/index.tsx',
	},
	output: {
		publicPath: '/_polo_/public/',
		filename: '[name].[fullhash].js',
		chunkFilename: '[name].[fullhash].js',
		path: path.resolve(cwd, 'pkg/services/static')
	},
	devtool: 'source-map',
	resolve: {
		extensions: ['.ts', '.tsx', '.js', '.scss', '.sass', '.hbs'],
		plugins: [new TsconfigPathsPlugin({ configFile: path.resolve(__dirname, 'tsconfig.json') })]
	},
	module: {
		rules: [
			{
				test: /\.tsx?$/,
				exclude: /node_modules/,
				use: [
					{
						loader: require.resolve('babel-loader'),
						options: {
							plugins: [
								isDevelopment && require.resolve('react-refresh/babel'),
							].filter(Boolean),
						},
					},
					{
						loader: 'ts-loader',
						options: {
							transpileOnly: true
						}
					},
				]

			},
			{
				test: /\.s[ac]ss$/,
				use: [
					isProduction ? MiniCssExtractPlugin.loader : 'style-loader',
					{
						loader: 'css-loader'
					},
					{
						loader: 'postcss-loader',
						options: {
							postcssOptions: {
								plugins: [
									require('postcss-import'),
									require('tailwindcss')(path.resolve(__dirname, './tailwind.config.js')),
									require('autoprefixer')
								]
							}
						}
					},
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
			{
				test: /\.css$/,
				use: [
					isProduction ? MiniCssExtractPlugin.loader : 'style-loader',
					{
						loader: 'css-loader'
					},
					{
						loader: 'postcss-loader',
						options: {
							postcssOptions: {
								plugins: [
									require('postcss-import'),
									require('tailwindcss')(path.resolve(__dirname, './tailwind.config.js')),
									require('autoprefixer')
								]
							}
						}
					},
				]
			},
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
		isDevelopment && new webpack.HotModuleReplacementPlugin(),
		isDevelopment && new ReactRefreshWebpackPlugin(),
		new ForkTsCheckerWebpackPlugin({
			typescript: {
				configFile: path.resolve(__dirname, 'tsconfig.json')
			}
		}),
		new MiniCssExtractPlugin({
			// Options similar to the same options in webpackOptions.output
			// both options are optional
			filename: "[name].[fullhash].css",
			chunkFilename: "[id].[fullhash].css",
		}),
		new HtmlWebpackPlugin({
			filename: './manager.html',
			template: path.resolve(__dirname, 'src', 'index.html'),
			chunks: ['manager'],
		}),
	].filter(Boolean),
	devServer: {
		static: {
			directory: path.join(__dirname, '../../public'),
		},
		compress: true,
		port: 9000,
		hot: true,
		allowedHosts: 'all'
	}
};