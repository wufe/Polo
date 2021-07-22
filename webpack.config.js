const manager = require('./packages/manager/webpack.config');
const helper = require('./packages/helper/webpack.config');

module.exports = [
    manager,
    helper
];