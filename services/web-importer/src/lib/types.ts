/**
 * Utility type that combines two types by taking all properties from T except those in K,
 * and then adding all properties from K.
 *
 * This is useful for extending a type with additional properties or overriding existing ones.
 */
export type With<T, K> = Omit<T, keyof K> & K;