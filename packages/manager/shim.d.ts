export { }
declare global {
    interface IProcess {
        env: {
            NODE_ENV: 'development' | 'production';
        }
    }

    interface IModule {
        hot: {
            accept(): void;
        }
    }
}