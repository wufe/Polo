import { IStore } from "./state/models/index";

declare module '*.jpg';
declare module '*.png';
declare module '*.svg';
declare module '*.webp';

export { }
declare global {
    interface Window {
        __REDUX_DEVTOOLS_EXTENSION__: Function;
        store: IStore;
    }

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