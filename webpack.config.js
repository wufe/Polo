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

module.exports = env => ({
    mode,
    entry: {
        'manager'       : './client/manager-index.tsx',
        'session-helper': './client/session-helper-index.tsx',
    },
    output: {
        publicPath: '/_polo_/static/',
        filename: '[name].js',
        path: path.resolve(__dirname, 'static')
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
                    MiniCssExtractPlugin.loader,
                    'css-loader',
                    'postcss-loader',
                    'sass-loader'
                ]
            },
            { test: /\.hbs$/, loader: "handlebars-loader" }

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
        isProduction && new CleanWebpackPlugin(),
        new ForkTsCheckerWebpackPlugin({
            typescript: {
                configFile: path.resolve(__dirname, 'tsconfig.json')
            }
        }),
        new MiniCssExtractPlugin({
            // Options similar to the same options in webpackOptions.output
            // both options are optional
            filename: "[name].css",
            chunkFilename: "[id].css",
        }),
        new HtmlWebpackPlugin({
            filename: './session-helper.html',
            template: './client/session-helper.hbs',
            inject: false,
            chunks: ['session-helper']
        }),
        new HtmlWebpackPlugin({
            filename: './manager.html',
            template: './client/manager.html',
            chunks: ['manager'],
        }),
        // isDevelopment && new webpack.HotModuleReplacementPlugin(),

    ].filter(Boolean),
    devServer: {
        contentBase: path.join(__dirname, 'static'),
        public: 'test.bembi.dev',
        compress: true,
        port: 9000,
        hot: true
    }
});