import { Instance, types } from "mobx-state-tree";
import { Service } from "./service";

enum SessionStatus {
    STARTING     = 'starting',
    STARTED      = 'started',
    START_FAILED = 'start_failed',
    STOPPING     = 'stopping',
}

export const Session = types.model({
    uuid      : types.string,
    name      : types.string,
    target    : types.string,
    port      : types.number,
    service   : Service,
    status    : types.enumeration<SessionStatus>(Object.values(SessionStatus)),
    logs      : types.array(types.string),
    checkout  : types.string,
    inactiveAt: types.string,
    folder    : types.string
});

export interface ISession extends Instance<typeof Session> {}