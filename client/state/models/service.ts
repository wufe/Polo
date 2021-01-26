import { Instance, SnapshotIn, SnapshotOut, types } from "mobx-state-tree";

export const Service = types.model({
    name: types.string,
    remote: types.string,
    target: types.string,
});

export interface IService extends Instance<typeof Service> { }
export interface IServiceSnapshotOut extends SnapshotOut<typeof Service> { }
export interface IServiceSnapshotIn extends SnapshotIn<typeof Service> { }