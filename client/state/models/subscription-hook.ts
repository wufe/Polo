import { store } from "."

export const useSubscription = () => {
    const { subscribeToSessionEvents, publishSessionEvent } = store.app;
    return {
        publish: publishSessionEvent,
        subscribe: subscribeToSessionEvents,
    };
}